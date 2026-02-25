package vault

import (
	"bytes"
	"testing"
)

func TestCompression_Integrity(t *testing.T) {
	data := []byte("The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog.")
	// Repeat to grow size
	for i := 0; i < 100; i++ {
		data = append(data, []byte("Repeat data...")...)
	}

	compBufPtr, framedSize, err := CompressBlock(data)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	defer ReleaseCompressed(compBufPtr)

	compBuf := (*compBufPtr)[:framedSize]

	uncompBufPtr, _, _, err := DecompressBlock(compBuf)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	defer ReleaseUncompressed(uncompBufPtr)

	uncompBuf := (*uncompBufPtr)[:len(data)]

	if !bytes.Equal(data, uncompBuf) {
		t.Fatal("decompressed data does not match original")
	}
}

func TestCompression_Empty(t *testing.T) {
	data := []byte("")
	compBufPtr, framedSize, err := CompressBlock(data)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	defer ReleaseCompressed(compBufPtr)

	compBuf := (*compBufPtr)[:framedSize]

	uncompBufPtr, _, _, err := DecompressBlock(compBuf)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	defer ReleaseUncompressed(uncompBufPtr)

	uncompBuf := (*uncompBufPtr)[:len(data)]
	if len(uncompBuf) != 0 {
		t.Fatal("expected empty decompressed data")
	}
}
