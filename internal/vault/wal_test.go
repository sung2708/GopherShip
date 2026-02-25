package vault

import (
	"os"
	"testing"

	"github.com/sungp/gophership/internal/buffer"
)

func TestWAL_Basic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gs-wal-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	w, err := NewWAL(tmpDir, 1024)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	data := buffer.MustAcquire(10)
	*data = append(*data, []byte("hello wal")...)

	w.MustWrite(data)
	w.Close()

	segments, err := w.ListSegments()
	if err != nil {
		t.Fatal(err)
	}

	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}

	// Verify content
	content, err := os.ReadFile(segments[0])
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "hello wal" {
		t.Errorf("expected 'hello wal', got '%s' (len=%d)", string(content), len(content))
	}
}

func TestWAL_Rotation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gs-wal-rot-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	segmentSize := int64(64)
	w, err := NewWAL(tmpDir, segmentSize)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Write 100 bytes (should trigger at least 1 rotation)
	for i := 0; i < 10; i++ {
		data := buffer.MustAcquire(10)
		*data = append(*data, []byte("1234567890")...)
		w.MustWrite(data)
	}

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

	segmentSize := int64(100)
	w, err := NewWAL(tmpDir, segmentSize)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Write 250 bytes (should span 3 segments)
	payload := make([]byte, 250)
	for i := range payload {
		payload[i] = 'A'
	}

	data := buffer.MustAcquire(250)
	*data = append(*data, payload...)
	w.MustWrite(data)
	w.Close()

	segments, err := w.ListSegments()
	if err != nil {
		t.Fatal(err)
	}

	if len(segments) != 3 {
		t.Errorf("expected 3 segments for 250 byte write with 100 byte capacity, got %d", len(segments))
	}

	// Verify total content size
	var totalSize int64
	for _, s := range segments {
		stat, _ := os.Stat(s)
		totalSize += stat.Size()
	}
	if totalSize != 250 {
		t.Errorf("expected total size 250, got %d", totalSize)
	}
}
