package ingester

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/sungp/gophership/internal/buffer"
)

// BenchmarkPivotLatency verifies that the ingestion reflex meets the < 500µs average target.
func TestIngester_AvgReflexLatency(t *testing.T) {
	// Setup ingester with small power-of-two buffer
	ing := NewIngester(2)
	ctx := context.Background()

	// Disable logging to avoid I/O bottlenecks during micro-benchmarking
	oldLevel := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.Disabled)
	defer zerolog.SetGlobalLevel(oldLevel)

	// Fill the buffer completely to trigger pivot on next calls
	data1 := buffer.MustAcquire(1024)
	ing.IngestData(ctx, data1)
	data2 := buffer.MustAcquire(1024)
	ing.IngestData(ctx, data2)

	const iterations = 1000000 // 1M iterations for robust statistical sampling

	// Track non-zero samples to verify timer resolution
	nonZeroSamples := 0
	// Create a large batch to measure average reflex time
	startBatch := time.Now()
	for i := 0; i < iterations; i++ {
		payload := buffer.MustAcquire(1024)
		opStart := time.Now()
		ing.IngestData(ctx, payload)
		if time.Since(opStart) > 0 {
			nonZeroSamples++
		}
	}
	totalElapsed := time.Since(startBatch)
	avgLatency := totalElapsed / iterations

	t.Logf("Total Time for %d reflex operations: %v", iterations, totalElapsed)
	t.Logf("Calculated Average Latency: %v", avgLatency)
	t.Logf("Timer Resolution Hits: %d/%d (%.4f%%)", nonZeroSamples, iterations, float64(nonZeroSamples)/float64(iterations)*100)

	if avgLatency > 500*time.Microsecond {
		t.Errorf("Average Reflex latency exceeded target: %v > 500µs", avgLatency)
	}
}

func BenchmarkPivotLatency(b *testing.B) {
	ing := NewIngester(2)
	ctx := context.Background()
	zerolog.SetGlobalLevel(zerolog.Disabled)

	data := buffer.MustAcquire(1024)
	ing.IngestData(ctx, data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		payload := buffer.MustAcquire(1024)
		ing.IngestData(ctx, payload)
	}
}

func TestIngester_ZeroAllocationPivot(t *testing.T) {
	ing := NewIngester(2)
	ctx := context.Background()
	zerolog.SetGlobalLevel(zerolog.Disabled)

	// Fill the buffer
	d1 := buffer.MustAcquire(64)
	ing.IngestData(ctx, d1)

	// Measure allocations for the pivot case
	allocs := testing.AllocsPerRun(100, func() {
		d2 := buffer.MustAcquire(64)
		ing.IngestData(ctx, d2)
	})

	if allocs > 0 {
		t.Errorf("expected 0 allocations on pivot path, got %v", allocs)
	}
}
