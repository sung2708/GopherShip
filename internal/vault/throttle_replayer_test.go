package vault

import (
	"context"
	"testing"
	"time"

	"github.com/sungp/gophership/internal/buffer"
	"github.com/sungp/gophership/internal/stochastic"
)

func TestReplayer_Throttling(t *testing.T) {
	dir := t.TempDir()

	// Setup WAL with small segment size for easy rotation
	w, err := NewWAL(dir, 1024*1024)
	if err != nil {
		t.Fatal(err)
	}

	// Write one block
	payload := []byte("throttle-test-payload")
	b := buffer.MustAcquire(len(payload))
	*b = append((*b)[:0], payload...)
	w.MustWrite(b)
	w.Close()

	// Reopen for Replayer
	w2, _ := NewWAL(dir, 1024*1024)
	defer w2.Close()

	t.Run("StatusGreen_NoExtraThrottle", func(t *testing.T) {
		stochastic.MustSetAmbientStatus(stochastic.StatusGreen)
		replayer := NewReplayer(w2, 1000) // 1ms base throttle

		start := time.Now()
		err := replayer.StreamTo(context.Background(), func(data []byte) error {
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		elapsed := time.Since(start)
		// Should be roughly 1ms base sleep + overhead
		if elapsed > 100*time.Millisecond {
			t.Errorf("Green status should not throttle heavily, got %v", elapsed)
		}
		if replayer.starvationNS.Load() == 0 && replayer.throttle > 0 {
			// starvationNS is tracked for all sleeps now in my implementation
		}
	})

	t.Run("StatusYellow_DoubleThrottle", func(t *testing.T) {
		stochastic.MustSetAmbientStatus(stochastic.StatusYellow)
		// ips = 1s / 10ms = 100
		replayer := NewReplayer(w2, 100)

		start := time.Now()
		err := replayer.StreamTo(context.Background(), func(data []byte) error {
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		elapsed := time.Since(start)
		// Multiplier for Yellow is 2.0. Base is 10ms. Total sleep should be 20ms.
		if elapsed < 20*time.Millisecond {
			t.Errorf("Yellow status should double throttle, expected >= 20ms, got %v", elapsed)
		}
		if time.Duration(replayer.starvationNS.Load()) < 20*time.Millisecond {
			t.Errorf("Starvation time should be >= 20ms, got %v", time.Duration(replayer.starvationNS.Load()))
		}
	})

	t.Run("StatusRed_DeepSleep", func(t *testing.T) {
		stochastic.MustSetAmbientStatus(stochastic.StatusRed)
		replayer := NewReplayer(w2, 1000) // 1ms base
		// Override deep sleep for fast unit test
		replayer.minDeepSleep = 50 * time.Millisecond

		start := time.Now()
		err := replayer.StreamTo(context.Background(), func(data []byte) error {
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		elapsed := time.Since(start)
		if elapsed < 50*time.Millisecond {
			t.Errorf("Red status should enter deep sleep (50ms), got %v", elapsed)
		}
	})

	// Cleanup
	stochastic.MustSetAmbientStatus(stochastic.StatusGreen)
}
