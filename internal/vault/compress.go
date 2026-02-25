package vault

import (
	"encoding/binary"
	"fmt"
	"hash"
	"hash/crc32"
	"sync"

	"github.com/pierrec/lz4/v4"
)

const (
	// WAL_MAGIC_LZ4 is "VLZ4" in big endian.
	WAL_MAGIC_LZ4 uint32 = 0x564C5A34
	// DefaultBlockSize is 64KB for optimal LZ4 performance.
	DefaultBlockSize = 64 * 1024
	// HeaderSize is Magic(4) + UncompressedLen(4) + CompressedLen(4) + CRC32(4).
	HeaderSize = 16
)

var (
	// blockPool manages uncompressed data blocks.
	blockPool = sync.Pool{
		New: func() interface{} {
			b := make([]byte, DefaultBlockSize)
			return &b
		},
	}

	// compPool manages compressed data blocks + header space.
	// We allocate extra space for the header and potential expansion.
	compPool = sync.Pool{
		New: func() interface{} {
			b := make([]byte, lz4.CompressBlockBound(DefaultBlockSize)+HeaderSize)
			return &b
		},
	}

	// crcPool manages CRC32 hashers.
	crcPool = sync.Pool{
		New: func() interface{} {
			return crc32.NewIEEE()
		},
	}
)

// CompressBlock compresses data into the framing format.
// It returns a pointer to the compressed buffer which MUST be returned to compPool.
func CompressBlock(data []byte) (*[]byte, int, error) {
	if len(data) > DefaultBlockSize {
		return nil, 0, fmt.Errorf("data exceeds block size")
	}

	compBufPtr := compPool.Get().(*[]byte)
	compBuf := *compBufPtr

	// Write Header: [Magic:4][UncompressedLen:4][CompressedLen:4]
	binary.BigEndian.PutUint32(compBuf[0:4], WAL_MAGIC_LZ4)
	binary.BigEndian.PutUint32(compBuf[4:8], uint32(len(data)))

	// Compress data into compBuf starting after header
	compressedSize, err := lz4.CompressBlock(data, compBuf[HeaderSize:], nil)
	if err != nil {
		compPool.Put(compBufPtr)
		return nil, 0, err
	}

	binary.BigEndian.PutUint32(compBuf[8:12], uint32(compressedSize))

	// Compute CRC32 of compressed data
	h := crcPool.Get().(hash.Hash32)
	h.Reset()
	h.Write(compBuf[HeaderSize : HeaderSize+compressedSize])
	checksum := h.Sum32()
	crcPool.Put(h)

	binary.BigEndian.PutUint32(compBuf[12:16], checksum)

	return compBufPtr, compressedSize + HeaderSize, nil
}

// DecompressBlock extracts data from the framing format.
// It returns a pointer to the uncompressed buffer, the total frame size (header+comp),
// the uncompressed size, and any error.
func DecompressBlock(framed []byte) (bufPtr *[]byte, totalSize int, uncompLen int, err error) {
	if len(framed) < HeaderSize {
		return nil, 0, 0, fmt.Errorf("framed data too short")
	}

	magic := binary.BigEndian.Uint32(framed[0:4])
	if magic != WAL_MAGIC_LZ4 {
		return nil, 0, 0, fmt.Errorf("invalid WAL magic: %x", magic)
	}

	uncompressedLen := binary.BigEndian.Uint32(framed[4:8])
	compressedLen := binary.BigEndian.Uint32(framed[8:12])
	headerChecksum := binary.BigEndian.Uint32(framed[12:16])

	if len(framed) < int(HeaderSize+compressedLen) {
		return nil, 0, 0, fmt.Errorf("truncated framed data")
	}

	// Verify CRC32
	h := crcPool.Get().(hash.Hash32)
	h.Reset()
	h.Write(framed[HeaderSize : HeaderSize+compressedLen])
	computedChecksum := h.Sum32()
	crcPool.Put(h)

	if computedChecksum != headerChecksum {
		return nil, 0, 0, fmt.Errorf("checksum mismatch: got %x, expected %x", computedChecksum, headerChecksum)
	}

	uncompBufPtr := blockPool.Get().(*[]byte)
	uncompBuf := *uncompBufPtr

	if uncompressedLen > uint32(len(uncompBuf)) {
		blockPool.Put(uncompBufPtr)
		return nil, 0, 0, fmt.Errorf("block exceeds buffer capacity")
	}

	n, err := lz4.UncompressBlock(framed[HeaderSize:HeaderSize+compressedLen], uncompBuf[:uncompressedLen])
	if err != nil {
		blockPool.Put(uncompBufPtr)
		return nil, 0, 0, err
	}

	if uint32(n) != uncompressedLen {
		blockPool.Put(uncompBufPtr)
		return nil, 0, 0, fmt.Errorf("decompression size mismatch: got %d, expected %d", n, uncompressedLen)
	}

	return uncompBufPtr, int(HeaderSize + compressedLen), int(uncompressedLen), nil
}

// ReleaseUncompressed returns a buffer to the blockPool.
func ReleaseUncompressed(b *[]byte) {
	blockPool.Put(b)
}

// ReleaseCompressed returns a buffer to the compPool.
func ReleaseCompressed(b *[]byte) {
	compPool.Put(b)
}
