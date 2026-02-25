package stochastic

import (
	"sync"
	"sync/atomic"
)

// AmbientStatus represents the global system pressure state (Green, Yellow, Red).
type AmbientStatus uint32

const (
	StatusGreen  AmbientStatus = 0 // Optimal: Full processing
	StatusYellow AmbientStatus = 1 // Under Load: Partial processing (skip enrichment)
	StatusRed    AmbientStatus = 2 // Critical: Raw byte flushing (Store Raw)
)

func (s AmbientStatus) String() string {
	switch s {
	case StatusGreen:
		return "GREEN"
	case StatusYellow:
		return "YELLOW"
	case StatusRed:
		return "RED"
	default:
		return "UNKNOWN"
	}
}

// globalState is the "Lazy Atomic" counter updated by an observer goroutine.
var globalState uint32

// stateMu protects the Prometheus Gauge and status listeners.
var stateMu sync.Mutex

// statusListeners stores channels for event-driven updates (Watch).
var statusListeners []chan AmbientStatus

// Monitor is the global host and component sensor.
var Monitor *SensingMonitor

// GetAmbientStatus returns the current global pressure state.
func GetAmbientStatus() AmbientStatus {
	return AmbientStatus(atomic.LoadUint32(&globalState))
}

// ThrottleMultiplier returns a sleep factor based on the current ambient status.
// - StatusGreen: 1.0 (Normal)
// - StatusYellow: 2.0 (Throttled)
// - StatusRed: 50.0 (Deep Sleep/Suspended)
func ThrottleMultiplier() float64 {
	status := GetAmbientStatus()
	switch status {
	case StatusYellow:
		return 2.0
	case StatusRed:
		return 50.0 // Effective suspension for background tasks
	default:
		return 1.0
	}
}

// MustSetAmbientStatus updates the global pressure state. Matches hot path naming.
func MustSetAmbientStatus(status AmbientStatus) {
	stateMu.Lock()
	defer stateMu.Unlock()

	prev := atomic.SwapUint32(&globalState, uint32(status))
	if uint32(status) != prev {
		IngesterZone.Set(float64(status))
		// Notify listeners of the state transition
		for _, ch := range statusListeners {
			select {
			case ch <- status:
			default: // Non-blocking to prevent slow listeners from hanging core reflexes
			}
		}
	}
}

// SubscribeStatus returns a channel that receives updates whenever the ambient status changes.
func SubscribeStatus() (chan AmbientStatus, func()) {
	stateMu.Lock()
	defer stateMu.Unlock()

	ch := make(chan AmbientStatus, 1)
	statusListeners = append(statusListeners, ch)

	cleanup := func() {
		stateMu.Lock()
		defer stateMu.Unlock()
		for i, l := range statusListeners {
			if l == ch {
				statusListeners = append(statusListeners[:i], statusListeners[i+1:]...)
				close(ch)
				break
			}
		}
	}

	return ch, cleanup
}

// SetGlobalMonitor initializes the global monitoring instance.
func SetGlobalMonitor(m *SensingMonitor) {
	stateMu.Lock()
	defer stateMu.Unlock()
	Monitor = m
}

// IncrementPressure increases the global pressure state (up to Red).
func IncrementPressure() {
	stateMu.Lock()
	defer stateMu.Unlock()

	curr := atomic.LoadUint32(&globalState)
	if curr < uint32(StatusRed) {
		next := curr + 1
		atomic.StoreUint32(&globalState, next)
		IngesterZone.Set(float64(next))
	}
}

// DecrementPressure decreases the global pressure state (down to Green).
func DecrementPressure() {
	stateMu.Lock()
	defer stateMu.Unlock()

	curr := atomic.LoadUint32(&globalState)
	if curr > uint32(StatusGreen) {
		next := curr - 1
		atomic.StoreUint32(&globalState, next)
		IngesterZone.Set(float64(next))
	}
}
