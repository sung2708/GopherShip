package stochastic

import (
	"testing"
)

func TestThrottleMultiplier(t *testing.T) {
	// Reset state
	MustSetAmbientStatus(StatusGreen)

	if m := ThrottleMultiplier(); m != 1.0 {
		t.Errorf("Expected 1.0 for Green, got %f", m)
	}

	MustSetAmbientStatus(StatusYellow)
	if m := ThrottleMultiplier(); m != 2.0 {
		t.Errorf("Expected 2.0 for Yellow, got %f", m)
	}

	MustSetAmbientStatus(StatusRed)
	if m := ThrottleMultiplier(); m != 50.0 {
		t.Errorf("Expected 50.0 for Red, got %f", m)
	}

	// Cleanup
	MustSetAmbientStatus(StatusGreen)
}
