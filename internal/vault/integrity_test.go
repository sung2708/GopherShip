package vault

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sungp/gophership/internal/buffer"
)

func TestIntegrity_Corruption(t *testing.T) {
	dir := t.TempDir()

	// 1. Setup WAL
	wal, err := NewWAL(dir, 1024*1024)
	if err != nil {
		t.Fatalf("failed to create WAL: %v", err)
	}

	// 2. Write a block
	payload := bytes.Repeat([]byte("integrity-test-data-block-"), 4000) // ~100KB > 64KB block
	bufPtr := buffer.MustAcquire(len(payload))
	*bufPtr = (*bufPtr)[:len(payload)] // RESLICE TO LENGTH!
	copy(*bufPtr, payload)
	wal.MustWrite(bufPtr)

	wal.Close()

	// 3. Find the segment file
	files, _ := os.ReadDir(dir)
	var segmentPath string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".log") {
			info, _ := f.Info()
			t.Logf("Found segment: %s, size: %d", f.Name(), info.Size())
			if info.Size() > 0 {
				segmentPath = filepath.Join(dir, f.Name())
			}
		}
	}
	if segmentPath == "" {
		t.Fatal("no non-empty segment file found")
	}

	// 4. Verify valid replay first
	wal2, _ := NewWAL(dir, 1024*1024)
	defer wal2.Close() // Ensure closed for Windows cleanup
	replayer := NewReplayer(wal2, 0)
	count := 0
	err = replayer.StreamTo(context.Background(), func(b []byte) error {
		count++
		t.Logf("Replayed block: %d, len: %d", count, len(b))
		return nil
	})
	if err != nil {
		t.Fatalf("replay of valid segment failed: %v", err)
	}
	if count == 0 {
		t.Errorf("expected at least 1 block, got 0")
	}

	// 5. Corrupt the file
	data, err := os.ReadFile(segmentPath)
	if err != nil {
		t.Fatalf("failed to read segment: %v", err)
	}

	// Flip a bit in the data portion (after 16-byte header)
	if len(data) > 20 {
		data[20] ^= 0xFF
		t.Logf("Corrupting file at offset 20 (size %d)", len(data))
	} else {
		t.Fatalf("segment too short to corrupt: %d bytes", len(data))
	}

	if err := os.WriteFile(segmentPath, data, 0644); err != nil {
		t.Fatalf("failed to write corrupted segment: %v", err)
	}

	// 6. Replay and verify mismatch error
	err = replayer.StreamTo(context.Background(), func(b []byte) error {
		return nil
	})

	if err == nil {
		t.Error("expected error on corrupted segment, got nil")
	} else {
		t.Logf("Got expected error: %v", err)
		if !strings.Contains(err.Error(), "checksum mismatch") {
			t.Errorf("expected checksum mismatch error, got: %v", err)
		}
	}
}

func BenchmarkIntegrity_Decompress(b *testing.B) {
	payload := make([]byte, 1024)
	compPtr, _, _ := CompressBlock(payload)
	framed := *compPtr

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		uncompPtr, _, _, err := DecompressBlock(framed)

		if err != nil {
			b.Fatalf("decompress failed: %v", err)
		}
		ReleaseUncompressed(uncompPtr)
	}
}
