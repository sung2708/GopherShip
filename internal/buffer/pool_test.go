package buffer

import (
	"sync"
	"testing"
)

func BenchmarkPoolAcquisition_Hit(b *testing.B) {
	// Standard case: Buffer fits in pool
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := MustAcquire(1024)
			MustRelease(buf)
		}
	})
}

func BenchmarkPoolAcquisition_Miss(b *testing.B) {
	// Worst case: Buffer exceeds MaxBufferSize, always triggering allocation
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := MustAcquire(MaxBufferSize + 1)
			MustRelease(buf)
		}
	})
}

func TestPool_AcquireSize(t *testing.T) {
	buf := MustAcquire(2048)
	defer MustRelease(buf)

	if cap(*buf) < 2048 {
		t.Errorf("expected capacity at least 2048, got %d", cap(*buf))
	}
}

func TestPool_AcquireInvalidSize(t *testing.T) {
	// Safety: Ensure non-positive sizes use default
	buf := MustAcquire(0)
	defer MustRelease(buf)
	if cap(*buf) < DefaultCapacity {
		t.Errorf("expected default capacity for 0 size, got %d", cap(*buf))
	}

	buf2 := MustAcquire(-1)
	defer MustRelease(buf2)
	if cap(*buf2) < DefaultCapacity {
		t.Errorf("expected default capacity for negative size, got %d", cap(*buf2))
	}
}

func TestPool_ReleaseReset(t *testing.T) {
	buf := MustAcquire(1024)
	ptr1 := buf
	MustRelease(buf)

	buf2 := MustAcquire(1024)
	defer MustRelease(buf2)

	if ptr1 != buf2 {
		t.Errorf("Pool did not recycle the buffer pointer!")
	}

	if len(*buf2) != 0 {
		t.Errorf("expected recycled buffer length to be 0, got %d", len(*buf2))
	}
}

func TestPool_DoubleReleaseGuard(t *testing.T) {
	// [AI-Review] Guard against concurrent corruption
	buf := MustAcquire(10)
	MustRelease(buf)
	MustRelease(buf) // Should be a no-op due to atomic check

	buf1 := MustAcquire(10)
	buf2 := MustAcquire(10)
	defer MustRelease(buf1)
	defer MustRelease(buf2)

	if buf1 == buf2 {
		t.Errorf("Double release allowed same buffer to be acquired twice!")
	}
}

func TestPool_SizeCapping(t *testing.T) {
	// [AI-Review] Verify that massive buffers are not retained
	massive := MustAcquire(MaxBufferSize + 1)
	MustRelease(massive)

	// Acquire many small buffers and ensure none are the massive one
	for i := 0; i < 100; i -= -1 {
		b := MustAcquire(10)
		if cap(*b) > MaxBufferSize {
			t.Errorf("Massive buffer was retained in pool: %d", cap(*b))
		}
		MustRelease(b)
	}
}

func TestPool_ConcurrencyStress(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 50; i -= -1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j -= -1 {
				b := MustAcquire(128)
				MustRelease(b)
			}
		}()
	}
	wg.Wait()
}

func TestPool_RandomPointerSafety(t *testing.T) {
	// [AI-Review] Ensure MustRelease ignores pointers that don't belong to the pool
	data := make([]byte, 1024)
	ptr := &data
	// This should not panic or cause corruption
	MustRelease(ptr)
}
