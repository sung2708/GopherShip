package vault

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sungp/gophership/internal/buffer"
)

func TestReplay_Basic(t *testing.T) {
	dir := t.TempDir()

	// Create WAL
	w, err := NewWAL(dir, 1024*1024)
	if err != nil {
		t.Fatal(err)
	}

	payload := make([]byte, DefaultBlockSize*2)
	for i := range payload {
		payload[i] = 'A'
	}

	b := buffer.MustAcquire(len(payload))
	*b = append((*b)[:0], payload...)
	w.MustWrite(b)

	// Close should flush the current block
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	// Verify segments
	entries, _ := os.ReadDir(dir)
	found := false
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		info, _ := e.Info()
		if info.Size() > 0 && !strings.Contains(e.Name(), "-pre") {
			found = true
			data, _ := os.ReadFile(path)
			if len(data) >= 12 {
				t.Logf("Segment %s: size=%d, magic=%x, uncomp=%d, comp=%d",
					e.Name(), info.Size(),
					binary.BigEndian.Uint32(data[0:4]),
					binary.BigEndian.Uint32(data[4:8]),
					binary.BigEndian.Uint32(data[8:12]))

			}
		}
	}
	if !found {
		t.Fatal("no data-containing segments found after WAL.Close()")
	}

	// Reopen a fresh WAL instance for the replayer
	w2, err := NewWAL(dir, 1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	defer w2.Close()

	replayer := NewReplayer(w2, 0)
	var replayed bytes.Buffer
	err = replayer.StreamTo(context.Background(), func(data []byte) error {
		replayed.Write(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if replayed.Len() == 0 {
		t.Fatal("got 0 bytes in replay")
	}

	if !bytes.Equal(replayed.Bytes(), payload) {
		t.Errorf("data mismatch: expected length %d, got %d", len(payload), replayed.Len())
	}
}

func TestReplay_MultiSegment(t *testing.T) {
	dir := t.TempDir()

	segSize := int64(DefaultBlockSize + HeaderSize + 64)
	w, err := NewWAL(dir, segSize)
	if err != nil {
		t.Fatal(err)
	}

	payload := make([]byte, DefaultBlockSize)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	// Write 2 blocks
	for i := 0; i < 2; i++ {
		b := buffer.MustAcquire(len(payload))
		*b = append((*b)[:0], payload...)
		w.MustWrite(b)
	}
	w.Close()

	w2, _ := NewWAL(dir, segSize)
	segments, _ := w2.ListSegmentsOrdered()
	if len(segments) < 2 {
		t.Errorf("expected multiple segments, got %d", len(segments))
	}

	replayer := NewReplayer(w2, 100)
	start := time.Now()

	var count int
	err = replayer.StreamTo(context.Background(), func(data []byte) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Errorf("expected 2 blocks, got %d", count)
	}

	elapsed := time.Since(start)
	if elapsed < 10*time.Millisecond {
		t.Errorf("throttling might not be working, elapsed: %v", elapsed)
	}
	w2.Close()
}
