package control

import (
	"context"
	"testing"

	"github.com/sungp/gophership/internal/stochastic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestHealthCheck_StatusMapping(t *testing.T) {
	// Reset status after test to prevent pollution (Architecture Mandate)
	oldStatus := stochastic.GetAmbientStatus()
	t.Cleanup(func() {
		stochastic.MustSetAmbientStatus(oldStatus)
	})

	s := NewServer("", "", nil, nil)

	// Test Green -> SERVING
	stochastic.MustSetAmbientStatus(stochastic.StatusGreen)
	resp, err := s.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Errorf("Expected SERVING for Green zone, got %v", resp.Status)
	}

	// Test Yellow -> SERVING (Optimistic, still serves but with warning pressure)
	stochastic.MustSetAmbientStatus(stochastic.StatusYellow)
	resp, _ = s.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Errorf("Expected SERVING for Yellow zone, got %v", resp.Status)
	}

	// Test Red -> NOT_SERVING (Pivoted to Vault, core reflex saturated)
	stochastic.MustSetAmbientStatus(stochastic.StatusRed)
	resp, _ = s.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	if resp.Status != grpc_health_v1.HealthCheckResponse_NOT_SERVING {
		t.Errorf("Expected NOT_SERVING for Red zone, got %v", resp.Status)
	}
}

func TestHealthCheck_TCP_Insecure_Access(t *testing.T) {
	port := "9094"
	s := NewServer(port, "/tmp/health_test.sock", nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := s.Start(ctx); err != nil {
		t.Fatalf("Server start failed: %v", err)
	}
	defer s.Stop(ctx)

	// Dial via TCP insecurely (Simulating K8s probe)
	conn, err := grpc.Dial("localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("Health check over insecure TCP failed: %v", err)
	}

	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Errorf("Expected SERVING, got %v", resp.Status)
	}
}
