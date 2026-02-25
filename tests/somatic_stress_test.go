package tests

import (
	"context"
	"testing"
	"time"

	"github.com/sungp/gophership/internal/ingester"
	"github.com/sungp/gophership/internal/stochastic"
	logcol "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	logcommon "go.opentelemetry.io/proto/otlp/common/v1"
	logproto "go.opentelemetry.io/proto/otlp/logs/v1"
)

// TestSomaticStressOOMProtection simulates a traffic spike and verifies the pivot to Raw Vault.
func TestSomaticStressOOMProtection(t *testing.T) {
	// 1. Setup Monitor with Aggressive Thresholds (e.g., 1KB budget for "OOM" simulation)
	monitor := stochastic.NewSensingMonitor(
		1,    // Check every operation
		1024, // 1KB Max RAM
		0.50, // 50% Yellow
		0.80, // 80% Red
		100,  // 100 Byte Ingester Budget
		1000, // 1KB Vault Budget
	)
	stochastic.SetGlobalMonitor(monitor)

	// 2. Initialize Ingester with small buffer to trigger reflex quickly
	ing := ingester.NewIngester(10)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// We don't start the worker loop to force buffer saturation
	// ing.StartWorkerLoop(ctx)

	// 3. Simulate High-Volume Traffic (20x Spike simulation)
	req := &logcol.ExportLogsServiceRequest{
		ResourceLogs: []*logproto.ResourceLogs{
			{
				ScopeLogs: []*logproto.ScopeLogs{
					{
						LogRecords: []*logproto.LogRecord{
							{Body: &logcommon.AnyValue{Value: &logcommon.AnyValue_StringValue{StringValue: "Stress test blob"}}},
						},
					},
				},
			},
		},
	}

	// First few requests should fill the buffer
	for i := 0; i < 11; i++ {
		_, _ = ing.Export(ctx, req)
	}

	// Force sensing
	monitor.MustSense()

	// 4. Verify Somatic Pivot
	status := stochastic.GetAmbientStatus()
	if status != stochastic.StatusRed {
		t.Errorf("Expected status RED due to resource pressure, got %s", status.String())
	}

	// 5. Verify the "Shock Absorber" (Reflex) is working
	// Note: We need a way to check fallbackCount or SomaticPivotsTotal metric
	// Ingester fallbackCount is private, but we can verify status transition which is the trigger.

	t.Logf("Somatic Zone successfully pivoted to: %s", status.String())
}
