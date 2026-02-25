package control

import (
	"context"
	"testing"

	"github.com/sungp/gophership/internal/somatic"
	"github.com/sungp/gophership/pkg/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockPressureProvider struct{}

func (m *mockPressureProvider) BufferDepth() int { return 0 }
func (m *mockPressureProvider) BufferCap() int   { return 100 }

func TestServer_OverrideSomaticZone(t *testing.T) {
	// 1. Setup Somatic Controller
	sc := somatic.NewController(&mockPressureProvider{})

	// 2. Setup Control Server
	socketPath := "/tmp/gophership_test.sock"
	// On Windows, UDS might behave differently, so we use TCP for the test if needed,
	// but the Server.Start listens on both. We'll try to dial the TCP port.
	port := "9093"
	srv := NewServer(port, socketPath, nil, sc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.Start(ctx); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer srv.Stop(ctx)

	// 3. Setup Client
	conn, err := grpc.Dial("localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()
	client := protocol.NewControlServiceClient(conn)

	// 4. Verify Initial State
	resp, err := client.GetSomaticStatus(ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}
	if resp.Zone != protocol.SomaticZone_ZONE_GREEN {
		t.Errorf("expected initial zone Green, got %v", resp.Zone)
	}

	// 5. Trigger Override to Red
	_, err = client.OverrideSomaticZone(ctx, &protocol.OverrideSomaticZoneRequest{
		Zone: protocol.SomaticZone_ZONE_RED,
	})
	if err != nil {
		t.Fatalf("failed to override: %v", err)
	}

	// 6. Verify Overridden State
	resp, err = client.GetSomaticStatus(ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}
	if resp.Zone != protocol.SomaticZone_ZONE_RED {
		t.Errorf("expected zone Red after override, got %v", resp.Zone)
	}

	// 7. Clear Override
	_, err = client.OverrideSomaticZone(ctx, &protocol.OverrideSomaticZoneRequest{
		Zone: protocol.SomaticZone_ZONE_UNSPECIFIED,
	})
	if err != nil {
		t.Fatalf("failed to clear override: %v", err)
	}

	// 8. Verify State Restored (Wait for Reassess if needed, but GetSomaticStatus
	// in current implementation just returns globalStatus which Reassess sets)
	// Actually Reassess must be called. In the real engine, it's called in loops.
	sc.Reassess()
	resp, err = client.GetSomaticStatus(ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}
	if resp.Zone != protocol.SomaticZone_ZONE_GREEN {
		t.Errorf("expected zone Green after clearing, got %v", resp.Zone)
	}
}
