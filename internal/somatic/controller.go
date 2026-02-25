package somatic

import (
	"sync/atomic"

	"github.com/sungp/gophership/internal/stochastic"
)

// PressureProvider is an interface implemented by components that track
// buffer occupancy (e.g., the Ingester).
type PressureProvider interface {
	BufferDepth() int
	BufferCap() int
}

// Controller manages the "biological" transitions between pressure zones.
type Controller struct {
	provider PressureProvider
	curr     uint32 // Atomic stochastic.AmbientStatus
	// overrideZone stores the manual override state.
	// 0 = No override (use sensors)
	// 1 = Green, 2 = Yellow, 3 = Red
	overrideZone uint32
}

func NewController(p PressureProvider) *Controller {
	c := &Controller{
		provider: p,
	}
	atomic.StoreUint32(&c.curr, uint32(stochastic.StatusGreen))
	atomic.StoreUint32(&c.overrideZone, 0)
	return c
}

// Override manually forces the controller into a specific state.
func (c *Controller) Override(status stochastic.AmbientStatus) {
	atomic.StoreUint32(&c.overrideZone, uint32(status)+1)
	atomic.StoreUint32(&c.curr, uint32(status))
	stochastic.MustSetAmbientStatus(status)
}

// ClearOverride removes any manual override and immediately reassesses state.
func (c *Controller) ClearOverride() {
	atomic.StoreUint32(&c.overrideZone, 0)
	c.Reassess()
}

// Reassess evaluates system pressure and updates the global stochastic state.
// It implements hysteresis to prevent rapid status oscillations.
func (c *Controller) Reassess() stochastic.AmbientStatus {
	override := atomic.LoadUint32(&c.overrideZone)
	if override != 0 {
		return stochastic.AmbientStatus(override - 1)
	}

	depth := c.provider.BufferDepth()
	capacity := c.provider.BufferCap()
	if capacity == 0 {
		return stochastic.StatusGreen
	}

	curr := stochastic.AmbientStatus(atomic.LoadUint32(&c.curr))
	next := curr

	// [NFR.P1] Optimization: Use integer math to avoid float64 overhead.
	// High Watermark: 85% occupancy triggers emergency mode.
	if depth*100 > capacity*85 {
		next = stochastic.StatusRed
	} else if depth*100 < capacity*20 {
		// Low Watermark: Only recover to Green if we drop below 20%.
		next = stochastic.StatusGreen
	}

	if next != curr {
		atomic.StoreUint32(&c.curr, uint32(next))
		stochastic.MustSetAmbientStatus(next)
	}

	return next
}
