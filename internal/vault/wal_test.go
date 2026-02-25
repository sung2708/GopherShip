package vault

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"testing"

	"github.com/sungp/gophership/internal/buffer"
)

func DecompressWALSegment(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result []byte
	off := 0
	for off < len(content) {
		// Minimum frame size is HeaderSize (12)
		if len(content[off:]) < HeaderSize {
			break
		}

		compressedLen := binary.BigEndian.Uint32(content[off+8 : off+12])
		frameSize := int(HeaderSize + compressedLen)

		if len(content[off:]) < frameSize {
			return nil, fmt.Errorf("truncated frame at offset %d", off)
		}

		uncompBufPtr, _, _, err := DecompressBlock(content[off : off+frameSize])
		if err != nil {
			return nil, fmt.Errorf("decompression failed at offset %d: %w", off, err)
		}

		uncompBuf := *uncompBufPtr
		// We need to know the uncompressed size from the header
		uncompressedLen := binary.BigEndian.Uint32(content[off+4 : off+8])
		result = append(result, uncompBuf[:uncompressedLen]...)
		ReleaseUncompressed(uncompBufPtr)

		off += frameSize
	}
	return result, nil
}

func TestWAL_Basic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gs-wal-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	w, err := NewWAL(tmpDir, 1024*1024) // 1MB segment size
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	payload := "hello wal"
	data := buffer.MustAcquire(len(payload))
	*data = append(*data, []byte(payload)...)

	w.MustWrite(data)
	w.Close()

	segments, err := w.ListSegments()
	if err != nil {
		t.Fatal(err)
	}

	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}

	// Verify decompressed content
	content, err := DecompressWALSegment(segments[0])
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != payload {
		t.Errorf("expected '%s', got '%s' (len=%d)", payload, string(content), len(content))
	}
}

func TestWAL_Rotation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gs-wal-rot-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// small segment size to trigger rotation
	// Note: with compression, must be larger than a few framed blocks.
	// Force multiple block flushes by writing > 64KB
	// segmentSize 10KB, 200KB total data -> many segments
	segmentSize := int64(128 * 1024) // 128KB segment size
	w, err := NewWAL(tmpDir, segmentSize)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	for i := 0; i < 20; i++ {
		payload := make([]byte, 10*1024)
		rand.Read(payload)
		data := buffer.MustAcquire(len(payload))
		*data = append(*data, payload...)
		w.MustWrite(data)
	}
	w.Close()

	segments, err := w.ListSegments()
	if err != nil {
		t.Fatal(err)
	}

	if len(segments) < 2 {
		t.Errorf("expected multiple segments after rotation, got %d", len(segments))
	}
}

func BenchmarkWAL_MustWrite(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "gs-wal-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	w, err := NewWAL(tmpDir, 1024*1024*64) // 64MB
	if err != nil {
		b.Fatal(err)
	}
	defer w.Close()

	payload := []byte("high-performance-log-entry-with-some-length")
	size := len(payload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := buffer.MustAcquire(size)
		copy(*data, payload)
		*data = (*data)[:size]

		w.MustWrite(data)
	}
}

func TestWAL_LargeWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gs-wal-large-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Small segment size to trigger spanning
	segmentSize := int64(128 * 1024) // 128KB segment size
	w, err := NewWAL(tmpDir, segmentSize)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Write 200KB to trigger multiple blocks and rotations
	payloadSize := 200 * 1024
	payload := make([]byte, payloadSize)
	rand.Read(payload)
	data := buffer.MustAcquire(payloadSize)
	*data = append(*data, payload...)
	w.MustWrite(data)
	w.Close()

	segments, err := w.ListSegments()
	if err != nil {
		t.Fatal(err)
	}

	if len(segments) < 2 {
		t.Errorf("expected multiple segments for large write, got %d", len(segments))
	}

	// Verify total decompressed content
	var fullContent []byte
	for _, s := range segments {
		uncomp, err := DecompressWALSegment(s)
		if err != nil {
			t.Fatalf("failed to decompress %s: %v", s, err)
		}
		fullContent = append(fullContent, uncomp...)
	}

	if !bytes.Equal(fullContent, payload) {
		t.Errorf("decompressed content size mismatch: expected 1000, got %d", len(fullContent))
	}
}
