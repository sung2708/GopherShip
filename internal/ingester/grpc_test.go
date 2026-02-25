package ingester

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	logcol "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	logcommon "go.opentelemetry.io/proto/otlp/common/v1"
	loglogs "go.opentelemetry.io/proto/otlp/logs/v1"
	logresource "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestIngester_OTLPgRPCIngestion(t *testing.T) {
	// 1. Setup Ingester and gRPC Server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ing := NewIngester(10)
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	addr := lis.Addr().String()

	s := grpc.NewServer()
	// This will fail to compile if Ingester doesn't implement the interface
	logcol.RegisterLogsServiceServer(s, ing)

	go func() {
		if err := s.Serve(lis); err != nil {
			return
		}
	}()
	defer s.Stop()

	// 2. Setup Client
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()
	client := logcol.NewLogsServiceClient(conn)

	// [Hardening] Start worker loop to consume from buffer
	ing.StartWorkerLoop(ctx)

	// 3. Send Mock OTLP Request
	req := &logcol.ExportLogsServiceRequest{
		ResourceLogs: []*loglogs.ResourceLogs{
			{
				Resource: &logresource.Resource{
					Attributes: []*logcommon.KeyValue{
						{Key: "service.name", Value: &logcommon.AnyValue{Value: &logcommon.AnyValue_StringValue{StringValue: "test-service"}}},
					},
				},
			},
		},
	}
	resp, err := client.Export(ctx, req)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	// 4. Verify data was processed by the worker loop
	// We use a timeout to ensure we don't hang if processedCount isn't updated
	start := time.Now()
	for time.Since(start) < 2*time.Second {
		if atomic.LoadUint64(&ing.processedCount) > 0 {
			return // Success
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("worker loop did not process any data within timeout")
}
