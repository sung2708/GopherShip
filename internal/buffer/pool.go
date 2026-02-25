package buffer

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	// DefaultCapacity is the size for new buffers when the pool is empty.
	DefaultCapacity = 1024
	// MaxBufferSize is the limit for pooled buffers to prevent memory bloat.
	MaxBufferSize = 65536 // 64KB
)

const magicVal = 0x50484952 // "SHIP" (GopherShip)

// pooledBuffer wraps the byte slice to track ownership and prevent double-release.
type pooledBuffer struct {
	magic  uint32 // Validation marker
	inPool int32  // 1 if in pool, 0 if leased
	data   []byte // The actual buffer
}

// pool is a global sync.Pool for *pooledBuffer.
var pool = sync.Pool{
	New: func() any {
		return &pooledBuffer{
			magic:  magicVal,
			inPool: 0,
			data:   make([]byte, 0, DefaultCapacity),
		}
	},
}

// MustAcquire returns a pointer to a byte slice of at least the requested size.
// AC3: If the pooled buffer is too small, a new one is created.
func MustAcquire(size int) *[]byte {
	if size <= 0 {
		size = DefaultCapacity
	}

	// [NFR.P2] Optimization: Try a few times to find a buffer with enough capacity
	// to avoid "Recycling Thrash" where the pool is full of small buffers.
	var pb *pooledBuffer
	for i := 0; i < 3; i++ {
		val := pool.Get()
		if val == nil {
			break
		}
		pb = val.(*pooledBuffer)
		if cap(pb.data) >= size {
			break
		}
		// Too small, put it back and try again
		atomic.StoreInt32(&pb.inPool, 1)
		pool.Put(pb)
		pb = nil
	}

	if pb == nil {
		// Create a specific one for this request.
		pb = &pooledBuffer{
			magic:  magicVal,
			inPool: 0,
			data:   make([]byte, 0, size),
		}
		// fmt.Printf("ALLOC: %p\n", pb)
		return &pb.data
	}

	// Mark as leased (0)
	atomic.StoreInt32(&pb.inPool, 0)
	pb.data = pb.data[:0]
	return &pb.data
}

// MustRelease returns a buffer to the pool.
// AC4: Resets length to 0 before release.
func MustRelease(buf *[]byte) {
	if buf == nil {
		return
	}

	// NFR.P1: Use unsafe to get the parent *pooledBuffer from the *[]byte (&pb.data).
	// [Safety Invariant]: This works because MustAcquire always returns a pointer to
	// the 'data' field of a *pooledBuffer. We subtract the offset of the 'data' field
	// from the pointer to 'data' to reach the start of the 'pooledBuffer' struct.
	// Structure Layout:
	// [ inPool (4 bytes) ][ padding ][ data (24 bytes) ]
	// ptr(buf) points to 'data'. ptr(buf) - offset(data) points to start of struct.
	pb := (*pooledBuffer)(unsafe.Pointer(uintptr(unsafe.Pointer(buf)) - unsafe.Offsetof(pooledBuffer{}.data)))

	// CRITICAL: Ownership & Corruption Protection
	// We check the magic number to ensure this pointer actually points to a data field
	// within a pooledBuffer struct. This prevents random pointers from corrupting the pool.
	if pb.magic != magicVal {
		// This is NOT a pooled buffer! Log a critical error or panic in dev.
		return
	}

	// CRITICAL: Double-Release Protection
	// Only proceed if we successfully transition from leased (0) to pooled (1).
	if !atomic.CompareAndSwapInt32(&pb.inPool, 0, 1) {
		// Already in pool or already released!
		return
	}

	// HIGH: Prevent Memory Bloat
	// If the buffer has grown beyond our threshold, don't return it to the pool.
	// This prevents a single massive request from permanently increasing memory usage.
	if cap(pb.data) > MaxBufferSize {
		return // Drop massive buffers, let GC collect them
	}

	pb.data = pb.data[:0]
	// fmt.Printf("RELEASE: %p\n", pb)
	pool.Put(pb)
}
