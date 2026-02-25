package stochastic

import (
	"runtime"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// SensingMonitor provides a zero-allocation, lock-free operational counter
// to trigger environmental sensing every N operations.
type SensingMonitor struct {
	// Padding to prevent false sharing (cache-line bouncing) on the hot counter.
	_       [64]byte
	counter atomic.Uint64
	_       [64]byte

	limit uint64
	mask  uint64

	// Thresholds (0-100)
	yellowRAMPerc float64
	redRAMPerc    float64
	maxRAM        uint64

	// CPU State (Windows simplified)
	cpuLoad atomic.Uint64 // Multiplied by 100 for integer precision

	// Component Budgets (Bytes)
	ingesterBudget uint64
	vaultBudget    uint64

	// Pre-calculated thresholds for zero-cycle checks
	ingesterYellow uint64
	ingesterRed    uint64
	vaultYellow    uint64
	vaultRed       uint64

	// Component Usage (Bytes) - Atomic for zero-allocation tracking
	ingesterUsage atomic.Int64
	vaultUsage    atomic.Int64
}

// NewSensingMonitor creates a new monitor with the specified limits and thresholds.
// yellowPerc and redPerc should be between 0 and 1.0 (e.g. 0.80 for 80%).
func NewSensingMonitor(limit uint64, maxRAM uint64, yellowPerc, redPerc float64, ingesterBudget, vaultBudget uint64) *SensingMonitor {
	var mask uint64
	if (limit != 0) && ((limit & (limit - 1)) == 0) {
		mask = limit - 1
	}

	m := &SensingMonitor{
		limit:          limit,
		mask:           mask,
		maxRAM:         maxRAM,
		yellowRAMPerc:  yellowPerc,
		redRAMPerc:     redPerc,
		ingesterBudget: ingesterBudget,
		vaultBudget:    vaultBudget,
		ingesterYellow: uint64(float64(ingesterBudget) * 0.80),
		ingesterRed:    uint64(float64(ingesterBudget) * 0.95),
		vaultYellow:    uint64(float64(vaultBudget) * 0.80),
		vaultRed:       uint64(float64(vaultBudget) * 0.95),
	}

	// Start background CPU sampler for Windows
	go m.sampleCPU()

	return m
}

// ShouldCheck returns true if the environment should be sensed based on the operation count.
// It uses atomic increment to ensure zero lock contention in the hot path.
func (m *SensingMonitor) ShouldCheck() bool {
	val := m.counter.Add(1)

	if m.mask != 0 {
		return (val & m.mask) == 0
	}

	return (val % m.limit) == 0
}

// MustSense Environment performs the actual sensing and updates the global ambient status.
func (m *SensingMonitor) MustSense() {
	memStatus := m.checkMemory()
	cpuStatus := m.checkCPU()
	ingesterStatus := m.checkIngester()
	vaultStatus := m.checkVault()

	// Update Metrics (AC5)
	IngesterUsageBytes.Set(float64(m.ingesterUsage.Load()))
	VaultUsageBytes.Set(float64(m.vaultUsage.Load()))

	// Escalation logic: StatusRed > StatusYellow > StatusGreen
	finalStatus := StatusGreen
	reason := ""
	if memStatus == StatusRed || cpuStatus == StatusRed || ingesterStatus == StatusRed || vaultStatus == StatusRed {
		finalStatus = StatusRed
		reason = "Critical resource pressure"
	} else if memStatus == StatusYellow || cpuStatus == StatusYellow || ingesterStatus == StatusYellow || vaultStatus == StatusYellow {
		finalStatus = StatusYellow
		reason = "High resource pressure"
	}

	// Update global state if changed
	prev := GetAmbientStatus()
	if finalStatus != prev {
		log.Warn().
			Str("prev", prev.String()).
			Str("curr", finalStatus.String()).
			Str("reason", reason).
			Msg("Somatic zone transition triggered by environment sensing")
		MustSetAmbientStatus(finalStatus)
	}
}

func (m *SensingMonitor) checkMemory() AmbientStatus {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	usage := ms.Sys // Use Sys as a proxy for total process footprint
	redThreshold := uint64(float64(m.maxRAM) * m.redRAMPerc)
	yellowThreshold := uint64(float64(m.maxRAM) * m.yellowRAMPerc)

	if usage >= redThreshold {
		return StatusRed
	}
	if usage >= yellowThreshold {
		return StatusYellow
	}
	return StatusGreen
}

func (m *SensingMonitor) checkCPU() AmbientStatus {
	// Simple threshold check against our background sampler
	load := m.cpuLoad.Load()
	if load >= 90 {
		return StatusRed
	}
	if load >= 75 {
		return StatusYellow
	}
	return StatusGreen
}

func (m *SensingMonitor) sampleCPU() {
	// Enhanced CPU sampling model for Windows (Proxy via Goroutines)
	// While LoadAvg is unavailable natively, Goroutine count is a high-fidelity
	// signal for somatic pressure within a Go process.
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Heuristic: maps NumGoroutine to a 0-100 "load" score.
		// > 1000 goroutines starts impacting scheduler efficiency.
		count := runtime.NumGoroutine()
		score := uint64(count / 10)
		if score > 100 {
			score = 100
		}
		m.cpuLoad.Store(score)
	}
}

// ReportIngesterUsage updates the atomic usage counter for the ingester.
func (m *SensingMonitor) ReportIngesterUsage(delta int64) {
	m.ingesterUsage.Add(delta)
}

// ReportVaultUsage updates the atomic usage counter for the vault.
func (m *SensingMonitor) ReportVaultUsage(delta int64) {
	m.vaultUsage.Add(delta)
}

func (m *SensingMonitor) checkIngester() AmbientStatus {
	usage := uint64(m.ingesterUsage.Load())
	if m.ingesterBudget == 0 {
		return StatusGreen
	}

	if usage >= m.ingesterRed {
		return StatusRed
	}
	if usage >= m.ingesterYellow {
		return StatusYellow
	}
	return StatusGreen
}

func (m *SensingMonitor) checkVault() AmbientStatus {
	usage := uint64(m.vaultUsage.Load())
	if m.vaultBudget == 0 {
		return StatusGreen
	}

	if usage >= m.vaultRed {
		return StatusRed
	}
	if usage >= m.vaultYellow {
		return StatusYellow
	}
	return StatusGreen
}

// Telemetry returns the current resource consumption and pressure score.
func (m *SensingMonitor) Telemetry() (usage, heap uint64, score uint32) {
	usage = uint64(m.ingesterUsage.Load() + m.vaultUsage.Load())

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	heap = ms.HeapObjects

	score = uint32(m.cpuLoad.Load())
	return
}
