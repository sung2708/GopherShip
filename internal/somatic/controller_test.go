package somatic

import (
	"testing"

	"github.com/sungp/gophership/internal/stochastic"
)

type mockProvider struct {
	depth int
	cap   int
}

func (m *mockProvider) BufferDepth() int { return m.depth }
func (m *mockProvider) BufferCap() int   { return m.cap }

func TestController_Reassess(t *testing.T) {
	mock := &mockProvider{cap: 1000}
	c := NewController(mock)

	// Case 1: Green Zone (0%)
	mock.depth = 0
	status := c.Reassess()
	if status != stochastic.StatusGreen {
		t.Errorf("expected StatusGreen at 0%%, got %s", status)
	}

	// Case 2: Hit High Watermark (86%)
	mock.depth = 860
	status = c.Reassess()
	if status != stochastic.StatusRed {
		t.Errorf("expected StatusRed at 86%%, got %s", status)
	}

	// Case 3: Hysteresis - Stay Red at 50%
	mock.depth = 500
	status = c.Reassess()
	if status != stochastic.StatusRed {
		t.Errorf("expected StatusRed to persist at 50%% due to hysteresis, got %s", status)
	}

	// Case 4: Recover at Low Watermark (19%)
	mock.depth = 190
	status = c.Reassess()
	if status != stochastic.StatusGreen {
		t.Errorf("expected StatusGreen recovery at 19%%, got %s", status)
	}
}
