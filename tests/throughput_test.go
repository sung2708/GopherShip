package tests

import (
	"context"
	"testing"

	"github.com/sungp/gophership/internal/ingester"
	"github.com/sungp/gophership/internal/stochastic"
	logcol "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	logcommon "go.opentelemetry.io/proto/otlp/common/v1"
	logproto "go.opentelemetry.io/proto/otlp/logs/v1"
)

// BenchmarkIngestionReflex measures the speed of the Ingester.Export call.
// Target: ~59ns per operation.
// Requirement: 0 B/op (Zero Allocation).
func BenchmarkIngestionReflex(b *testing.B) {
	// Setup a standard ingester with a large buffer to avoid backpressure during the bench
	ing := ingester.NewIngester(1024 * 1024)
	ctx := context.Background()
	ing.StartWorkerLoop(ctx)
	defer ing.Stop()

	// Disable global monitor to focus purely on the hot path logic
	stochastic.SetGlobalMonitor(nil)

	// Create a dummy OTLP request
	req := &logcol.ExportLogsServiceRequest{
		ResourceLogs: []*logproto.ResourceLogs{
			{
				ScopeLogs: []*logproto.ScopeLogs{
					{
						LogRecords: []*logproto.LogRecord{
							{
								Body: &logcommon.AnyValue{
									Value: &logcommon.AnyValue_StringValue{
										StringValue: "Biological reflex test signal",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := ing.Export(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkIngestData measures the raw internal ingestion reflex (select-default).
func BenchmarkIngestData(b *testing.B) {
	ing := ingester.NewIngester(1024 * 1024)
	ctx := context.Background()

	// Prepare data
	data := make([]byte, 1024)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// We mock the pointer-to-pooled-buffer pattern
		d := &data
		ing.IngestData(ctx, d)
	}
}
