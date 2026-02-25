package somatic

import (
	"testing"

	"github.com/sungp/gophership/internal/stochastic"
)

type overrideMockProvider struct {
	depth int
	cap   int
}

func (m *overrideMockProvider) BufferDepth() int { return m.depth }
func (m *overrideMockProvider) BufferCap() int   { return m.cap }

func TestController_Override(t *testing.T) {
	mp := &overrideMockProvider{depth: 0, cap: 100}
	c := NewController(mp)

	// 1. Initial state should be Green
	if status := c.Reassess(); status != stochastic.StatusGreen {
		t.Errorf("expected initial status Green, got %v", status)
	}

	// 2. Force Red via Override
	c.Override(stochastic.StatusRed)
	if status := c.Reassess(); status != stochastic.StatusRed {
		t.Errorf("expected status Red after override, got %v", status)
	}

	// 3. Even with low pressure, override should persist
	mp.depth = 5
	if status := c.Reassess(); status != stochastic.StatusRed {
		t.Errorf("expected override status Red to persist with low pressure, got %v", status)
	}

	// 4. Force Yellow via Override
	c.Override(stochastic.StatusYellow)
	if status := c.Reassess(); status != stochastic.StatusYellow {
		t.Errorf("expected status Yellow after override change, got %v", status)
	}

	// 5. Clear Override - should restore normal logic (low pressure = Green)
	c.ClearOverride()
	if status := c.Reassess(); status != stochastic.StatusGreen {
		t.Errorf("expected status Green after clearing override, got %v", status)
	}

	// 6. High pressure should now trigger Red normally
	mp.depth = 95
	if status := c.Reassess(); status != stochastic.StatusRed {
		t.Errorf("expected status Red with high pressure, got %v", status)
	}
}
