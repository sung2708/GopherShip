package vault

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/edsrzf/mmap-go"
	"github.com/rs/zerolog/log"
	"github.com/sungp/gophership/internal/stochastic"
)

const (
	// DefaultMinDeepSleep is the minimum wait in StatusRed to yield to ingester.
	DefaultMinDeepSleep = 5 * time.Second
	// DefaultYellowBaseThrottle is the fallback sleep for StatusYellow if no IPS configured.
	DefaultYellowBaseThrottle = 100 * time.Millisecond
)

type Replayer struct {
	wal          *WAL
	throttle     time.Duration
	minDeepSleep time.Duration

	// Starvation Metrics (nanoseconds)
	starvationNS atomic.Int64
	processingNS atomic.Int64
}

func NewReplayer(w *WAL, ips int) *Replayer {
	var t time.Duration
	if ips > 0 {
		t = time.Second / time.Duration(ips)
	}
	return &Replayer{
		wal:          w,
		throttle:     t,
		minDeepSleep: DefaultMinDeepSleep,
	}
}

func (r *Replayer) StreamTo(ctx context.Context, sink func([]byte) error) error {
	segments, err := r.wal.ListSegmentsOrdered()
	if err != nil {
		return err
	}

	log.Debug().Int("count", len(segments)).Msg("Replayer starting: discovered segments")
	for _, path := range segments {
		if err := r.streamSegment(ctx, path, sink); err != nil {
			return err
		}
	}
	return nil
}

func (r *Replayer) streamSegment(ctx context.Context, path string, sink func([]byte) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.Size() == 0 {
		return nil
	}

	m, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		return err
	}
	defer m.Unmap()

	segmentStart := time.Now()
	offset := int64(0)
	for offset+int64(HeaderSize) <= int64(len(m)) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Optimization: Find first non-zero magic number efficiently
		if m[offset] == 0 {
			next := bytes.IndexFunc(m[offset:], func(r rune) bool { return r != 0 })
			if next == -1 {
				break // All zeros till end
			}
			offset += int64(next)
			if offset+int64(HeaderSize) > int64(len(m)) {
				break
			}
		}

		// Processing starts here
		procStart := time.Now()
		// DecompressBlock already validates magic, length, and CRC32
		uncompPtr, total, uncompLen, err := DecompressBlock(m[offset:])
		if err != nil {
			log.Error().Err(err).Int64("offset", offset).Str("path", path).Msg("CRITICAL: Integrity failure or decompression error - quarantining segment")
			return fmt.Errorf("integrity failure at %d in %s: %w", offset, path, err)
		}

		if err := sink((*uncompPtr)[:uncompLen]); err != nil {
			ReleaseUncompressed(uncompPtr)
			return err
		}
		ReleaseUncompressed(uncompPtr)
		r.processingNS.Add(time.Since(procStart).Nanoseconds())

		offset += int64(total)

		// Throttling logic based on Stochastic pressure (AC1, AC4)
		mult := stochastic.ThrottleMultiplier()
		wait := r.throttle

		if mult > 1.0 {
			wait = time.Duration(float64(r.throttle) * mult)
			// Ensure visibility of the throttle event (AC5)
			status := stochastic.GetAmbientStatus()

			// Base throttle for Yellow if none configured
			if r.throttle == 0 && status == stochastic.StatusYellow {
				wait = DefaultYellowBaseThrottle
			}

			// Deep Sleep for Red Zone (AC3)
			if status == stochastic.StatusRed && wait < r.minDeepSleep {
				wait = r.minDeepSleep
			}

			log.Warn().
				Str("status", status.String()).
				Float64("multiplier", mult).
				Dur("wait", wait).
				Dur("cumulative_starvation", time.Duration(r.starvationNS.Load())).
				Msg("Replayer yielding CPU due to system pressure")
		}

		if wait > 0 {
			t0 := time.Now()
			time.Sleep(wait)
			r.starvationNS.Add(time.Since(t0).Nanoseconds())
		}
	}

	// Report Starvation Score (AC5)
	totalTime := time.Since(segmentStart)
	score := 0.0
	procNS := r.processingNS.Load()
	starvNS := r.starvationNS.Load()

	if procNS > 0 {
		score = float64(starvNS) / float64(procNS)
	}

	log.Info().
		Str("path", path).
		Dur("processing_time", time.Duration(procNS)).
		Dur("starvation_time", time.Duration(starvNS)).
		Dur("total_segment_time", totalTime).
		Float64("starvation_score", score).
		Msg("Segment streaming complete")

	return nil
}
