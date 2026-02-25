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
)

const (
	// DefaultSegmentSize is 64MB as per Story 3.1.
	DefaultSegmentSize = 64 * 1024 * 1024
	// WALFilePrefix is the mandatory prefix for segment files.
	WALFilePrefix = "gs-vault-"
	// WALFileSuffix is the mandatory suffix for segment files.
	WALFileSuffix = ".log"
)

// Segment represents a single memory-mapped file in the WAL.
type Segment struct {
	file    *os.File
	mmap    mmap.MMap
	writeAt int64
	size    int64
	path    string
}

// WAL manages a directory of segments for high-performance durability.
type WAL struct {
	mu            sync.RWMutex
	dir           string
	segmentSize   int64
	activeSegment *Segment
	index         uint64 // AC4: Current segment index
	closed        bool

	prealloc  chan string
	closeOnce sync.Once
}

// NewWAL initializes a new Write-Ahead Log in the specified directory.
func NewWAL(dir string, segmentSize int64) (*WAL, error) {
	if segmentSize <= 0 {
		segmentSize = DefaultSegmentSize
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create WAL directory: %w", err)
	}

	w := &WAL{
		dir:         dir,
		segmentSize: segmentSize,
		prealloc:    make(chan string, 1),
	}

	// Scan directory to find the next index
	files, _ := os.ReadDir(dir)
	var maxIdx uint64
	for _, f := range files {
		if strings.HasPrefix(f.Name(), WALFilePrefix) {
			parts := strings.Split(strings.TrimSuffix(f.Name(), WALFileSuffix), "-")
			if len(parts) >= 3 {
				var idx uint64
				fmt.Sscanf(parts[len(parts)-1], "%d", &idx)
				if idx > maxIdx {
					maxIdx = idx
				}
			}
		}
	}
	w.index = maxIdx

	// Start pre-allocation worker
	go w.preallocWorker()

	// Initial segment
	if err := w.rotate(); err != nil {
		return nil, err
	}

	log.Info().Str("dir", dir).Int64("segment_size", segmentSize).Uint64("start_index", w.index).Msg("Hardened WAL initialized with pre-allocation")
	return w, nil
}

// preallocWorker prepares the next segment in the background to minimize rotation latency.
// [MED] Performance Optimization
func (w *WAL) preallocWorker() {
	for {
		w.mu.RLock()
		if w.closed {
			w.mu.RUnlock()
			return
		}
		dir := w.dir
		size := w.segmentSize
		nextIdx := w.index + 1
		w.mu.RUnlock()

		timestamp := time.Now().UnixNano() / int64(time.Millisecond)
		filename := fmt.Sprintf("%s%d-%d-pre%s", WALFilePrefix, timestamp, nextIdx, WALFileSuffix)
		path := filepath.Join(dir, filename)

		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Error().Err(err).Msg("Failed to pre-allocate WAL segment file")
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if err := f.Truncate(size); err != nil {
			log.Error().Err(err).Msg("Failed to truncate pre-allocated WAL segment file")
			f.Close()
			time.Sleep(100 * time.Millisecond)
			continue
		}
		f.Close()

		select {
		case w.prealloc <- path:
			// Pre-allocation successful
		case <-time.After(1 * time.Second):
			// WAL closed or pool full
			os.Remove(path)
			return
		}
	}
}

// MustWrite appends data to the WAL. It is zero-allocation and handles buffer release.
// AC5: Leveraging internal/buffer.MustRelease for cleanup.
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
		space := w.activeSegment.size - w.activeSegment.writeAt
		if space <= 0 {
			if err := w.rotate(); err != nil {
				log.Error().Err(err).Msg("Critical: Fail to rotate WAL; data loss")
				break
			}
			space = w.activeSegment.size - w.activeSegment.writeAt
		}

		chunkSize := int64(len(remaining))
		if chunkSize > space {
			chunkSize = space
		}

		// Copy chunk to memory map
		copy(w.activeSegment.mmap[w.activeSegment.writeAt:], remaining[:chunkSize])
		w.activeSegment.writeAt += chunkSize
		remaining = remaining[chunkSize:]
	}

	// Release the pooled buffer immediately after persistence
	buffer.MustRelease(data)
}

// rotate closes the current segment and opens/creates a new one.
// [MED] Performance Optimization: Pulls pre-created file from pre-allocation channel.
func (w *WAL) rotate() error {
	if w.activeSegment != nil {
		if err := w.activeSegment.close(); err != nil {
			log.Warn().Err(err).Str("path", w.activeSegment.path).Msg("Failed to close old WAL segment")
		}
	}

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	w.index++
	finalFilename := fmt.Sprintf("%s%d-%d%s", WALFilePrefix, timestamp, w.index, WALFileSuffix)
	finalPath := filepath.Join(w.dir, finalFilename)

	var nextPath string
	select {
	case nextPath = <-w.prealloc:
		// Rename pre-allocated file BEFORE mmapping (Windows compatibility)
		if err := os.Rename(nextPath, finalPath); err != nil {
			log.Warn().Err(err).Str("from", nextPath).Str("to", finalPath).Msg("Pre-allocation rename failed; fallback")
			os.Remove(nextPath)
		} else {
			f, err := os.OpenFile(finalPath, os.O_RDWR, 0644)
			if err != nil {
				return err
			}
			m, err := mmap.Map(f, mmap.RDWR, 0)
			if err != nil {
				f.Close()
				return err
			}
			w.activeSegment = &Segment{
				file: f,
				mmap: m,
				size: w.segmentSize,
				path: finalPath,
			}
			return nil
		}
	default:
		// Fall through to fallback
	}

	// Fallback to synchronous allocation
	f, err := os.OpenFile(finalPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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
	w.activeSegment = &Segment{
		file:    f,
		mmap:    m,
		size:    w.segmentSize,
		path:    finalPath,
		writeAt: 0,
	}
	return nil
}

// Close gracefully shuts down the WAL, flushing all mapped data to disk.
func (w *WAL) Close() error {
	var err error
	w.closeOnce.Do(func() {
		w.mu.Lock()
		w.closed = true
		w.mu.Unlock()

		// Drain prealloc channel and delete files
		closeWaiting := true
		for closeWaiting {
			select {
			case path := <-w.prealloc:
				os.Remove(path)
			default:
				closeWaiting = false
			}
		}

		w.mu.Lock()
		if w.activeSegment != nil {
			err = w.activeSegment.close()
		}
		w.mu.Unlock()
	})
	return err
}

// close unmaps and closes the segment file.
func (s *Segment) close() error {
	if s.mmap != nil {
		// Flush mapping to disk
		if err := s.mmap.Flush(); err != nil {
			log.Warn().Err(err).Str("path", s.path).Msg("Failed to flush mmap")
		}

		if err := s.mmap.Unmap(); err != nil {
			log.Warn().Err(err).Str("path", s.path).Msg("Failed to unmap")
		}
	}

	// Truncate to actual written size before closing
	_ = s.file.Truncate(s.writeAt)
	return s.file.Close()
}

// ListSegments returns a list of all WAL segments in the directory, sorted by timestamp.
func (w *WAL) ListSegments() ([]string, error) {
	files, err := os.ReadDir(w.dir)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, f := range files {
		if !f.IsDir() &&
			strings.HasPrefix(f.Name(), WALFilePrefix) &&
			strings.HasSuffix(f.Name(), WALFileSuffix) &&
			!strings.Contains(f.Name(), "-pre") {
			results = append(results, filepath.Join(w.dir, f.Name()))
		}
	}

	sort.Strings(results)
	return results, nil
}
