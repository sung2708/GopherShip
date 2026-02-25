package stochastic

import (
	"runtime"
	"testing"
)

func TestSensingMonitor(t *testing.T) {
	t.Run("PowerOfTwoLimit", func(t *testing.T) {
		limit := uint64(1024)
		monitor := NewSensingMonitor(limit, 1024*1024*1024, 0.80, 0.95, 100*1024*1024, 100*1024*1024)

		checks := 0
		for i := uint64(1); i <= limit*3; i++ {
			if monitor.ShouldCheck() {
				checks++
				if i%limit != 0 {
					t.Errorf("ShouldCheck() returned true at operation %d, expected multiple of %d", i, limit)
				}
			}
		}

		if checks != 3 {
			t.Errorf("Expected 3 checks for 3072 operations at limit 1024, but got %d", checks)
		}
	})
}

func TestSensingMonitor_Sense(t *testing.T) {
	// Initialize global state to Green
	MustSetAmbientStatus(StatusGreen)

	// Create monitor with very low RAM limit to trigger status change
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	// Set maxRAM to something very small (e.g., current usage / 2)
	maxRAM := ms.Sys / 2
	monitor := NewSensingMonitor(1024, maxRAM, 0.80, 0.95, 100*1024*1024, 100*1024*1024)

	monitor.MustSense()

	status := GetAmbientStatus()
	if status != StatusRed {
		t.Errorf("Expected StatusRed due to low RAM limit, got %s", status.String())
	}

	// Reset to a large RAM limit
	monitor.maxRAM = ms.Sys * 10

	monitor.MustSense()
	status = GetAmbientStatus()
	if status != StatusGreen {
		t.Errorf("Expected StatusGreen after resetting RAM limit, got %s", status.String())
	}
}

func TestSensingMonitor_ComponentBudgets(t *testing.T) {
	// Initialize global state to Green
	MustSetAmbientStatus(StatusGreen)

	limit := uint64(1)
	ingesterBudget := uint64(100 * 1024) // 100KB
	vaultBudget := uint64(100 * 1024)    // 100KB
	monitor := NewSensingMonitor(limit, 1024*1024*1024, 0.8, 0.9, ingesterBudget, vaultBudget)

	// Case 1: Ingester breach Red (96KB > 95KB)
	monitor.ReportIngesterUsage(96 * 1024)
	if monitor.ShouldCheck() {
		monitor.MustSense()
	}

	status := GetAmbientStatus()
	if status != StatusRed {
		t.Errorf("Expected StatusRed due to ingester budget breach, got %s", status.String())
	}

	// Reset
	MustSetAmbientStatus(StatusGreen)
	monitor.ReportIngesterUsage(-96 * 1024)

	// Case 2: Vault breach Yellow (85KB > 80KB)
	monitor.ReportVaultUsage(85 * 1024)
	if monitor.ShouldCheck() {
		monitor.MustSense()
	}

	status = GetAmbientStatus()
	if status != StatusYellow {
		t.Errorf("Expected StatusYellow due to vault budget breach, got %s", status.String())
	}
}

func BenchmarkSensingMonitor_ShouldCheck(t *testing.B) {
	monitor := NewSensingMonitor(1024, 1024*1024*1024, 0.80, 0.95, 100*1024*1024, 100*1024*1024)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		monitor.ShouldCheck()
	}
}
