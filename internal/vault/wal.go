package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/edsrzf/mmap-go"
	"github.com/rs/zerolog/log"
	"github.com/sungp/gophership/internal/buffer"
	"github.com/sungp/gophership/internal/stochastic"
)

const (
	DefaultSegmentSize = 64 * 1024 * 1024
	WALFilePrefix      = "gs-vault-"
	WALFileSuffix      = ".log"
)

type Segment struct {
	file    *os.File
	mmap    mmap.MMap
	writeAt int64
	size    int64
	path    string
}

type WAL struct {
	mu            sync.Mutex
	dir           string
	segmentSize   int64
	activeSegment *Segment
	index         uint64
	closed        bool

	currBlock    *[]byte
	currBlockOff int
}

func NewWAL(dir string, segmentSize int64) (*WAL, error) {
	if segmentSize < int64(DefaultBlockSize+HeaderSize) {
		segmentSize = DefaultSegmentSize
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create WAL dir: %w", err)
	}

	w := &WAL{
		dir:         dir,
		segmentSize: segmentSize,
	}

	// Simple index recovery
	entries, _ := os.ReadDir(dir)
	var maxIdx uint64
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), WALFilePrefix) && strings.HasSuffix(e.Name(), WALFileSuffix) {
			var ts int64
			var idx uint64
			fmt.Sscanf(e.Name(), WALFilePrefix+"%d-%d"+WALFileSuffix, &ts, &idx)
			if idx > maxIdx {
				maxIdx = idx
			}
		}
	}
	w.index = maxIdx

	if err := w.rotateLocked(); err != nil {
		return nil, fmt.Errorf("failed to rotate initial segment: %w", err)
	}

	w.currBlock = blockPool.Get().(*[]byte)
	w.currBlockOff = 0

	// Report initial usage
	if stochastic.Monitor != nil {
		stochastic.Monitor.ReportVaultUsage(int64(DefaultBlockSize))
	}

	log.Info().Str("dir", dir).Uint64("start_index", w.index).Msg("WAL initialized")
	return w, nil
}

func (w *WAL) MustWrite(data *[]byte) {
	if data == nil || len(*data) == 0 {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		buffer.MustRelease(data)
		return
	}

	remaining := *data
	for len(remaining) > 0 {
		space := DefaultBlockSize - w.currBlockOff

		if space == 0 {
			if err := w.flushBlockLocked(); err != nil {
				panic(err)
			}
			space = DefaultBlockSize
		}

		n := len(remaining)
		if n > space {
			n = space
		}

		copy((*w.currBlock)[w.currBlockOff:], remaining[:n])
		w.currBlockOff += n
		remaining = remaining[n:]

		if w.currBlockOff == DefaultBlockSize {
			if err := w.flushBlockLocked(); err != nil {
				panic(err)
			}
		}
	}
	buffer.MustRelease(data)
}

func (w *WAL) flushBlockLocked() error {
	if w.currBlockOff == 0 {
		return nil
	}

	compBufPtr, framedSize, err := CompressBlock((*w.currBlock)[:w.currBlockOff])
	if err != nil {
		return err
	}
	defer ReleaseCompressed(compBufPtr)

	needed := int64(framedSize)
	if w.activeSegment == nil || w.activeSegment.writeAt+needed > w.activeSegment.size {
		if err := w.rotateLocked(); err != nil {
			return err
		}
	}

	copy(w.activeSegment.mmap[w.activeSegment.writeAt:], (*compBufPtr)[:framedSize])
	log.Debug().Int("size", framedSize).Int64("at", w.activeSegment.writeAt).Msg("WAL block flushed")
	w.activeSegment.writeAt += needed
	w.currBlockOff = 0
	return nil
}

func (w *WAL) rotateLocked() error {
	if w.activeSegment != nil {
		if err := w.activeSegment.close(); err != nil {
			log.Error().Err(err).Msg("failed to close segment during rotation")
		}
	}

	w.index++
	path := filepath.Join(w.dir, fmt.Sprintf("%s%d-%d%s", WALFilePrefix, time.Now().UnixNano()/1e6, w.index, WALFileSuffix))

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	if err := f.Truncate(w.segmentSize); err != nil {
		f.Close()
		return err
	}
	m, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		f.Close()
		return err
	}
	w.activeSegment = &Segment{file: f, mmap: m, size: w.segmentSize, path: path, writeAt: 0}

	// Report usage increase
	if stochastic.Monitor != nil {
		stochastic.Monitor.ReportVaultUsage(w.segmentSize)
	}
	return nil
}

func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}
	w.closed = true
	if err := w.flushBlockLocked(); err != nil {
		log.Error().Err(err).Msg("failed to flush final block on close")
	}
	if w.activeSegment != nil {
		if err := w.activeSegment.close(); err != nil {
			log.Error().Err(err).Msg("failed to close active segment on close")
		}
	}
	if w.currBlock != nil {
		ReleaseUncompressed(w.currBlock)
		w.currBlock = nil
		// Report usage reduction
		if stochastic.Monitor != nil {
			stochastic.Monitor.ReportVaultUsage(-int64(DefaultBlockSize))
		}
	}
	return nil
}

func (s *Segment) close() error {
	_ = s.mmap.Flush()
	_ = s.mmap.Unmap()
	_ = s.file.Sync()
	_ = s.file.Truncate(s.writeAt)
	err := s.file.Close()

	// Report usage reduction
	if stochastic.Monitor != nil {
		stochastic.Monitor.ReportVaultUsage(-s.size)
	}
	return err
}

func (w *WAL) ListSegmentsOrdered() ([]string, error) {
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), WALFilePrefix) && strings.HasSuffix(e.Name(), WALFileSuffix) {
			res = append(res, filepath.Join(w.dir, e.Name()))
		}
	}
	sort.Strings(res)
	return res, nil
}

func (w *WAL) ListSegments() ([]string, error) {
	return w.ListSegmentsOrdered()
}
