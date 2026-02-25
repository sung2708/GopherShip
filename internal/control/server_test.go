package control

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/sungp/gophership/pkg/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestWatchSomaticStatus(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := NewServer("9092", "/tmp/gs-test.sock", nil, nil)

	grpcServer := grpc.NewServer()
	protocol.RegisterControlServiceServer(grpcServer, s)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			return
		}
	}()
	defer grpcServer.Stop()

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := protocol.NewControlServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stream, err := client.WatchSomaticStatus(ctx, &protocol.WatchStatusRequest{
		RefreshIntervalMs: 500,
	})
	if err != nil {
		t.Fatalf("Failed to start stream: %v", err)
	}

	// Receive at least two updates
	for i := 0; i < 2; i++ {
		resp, err := stream.Recv()
		if err != nil {
			t.Fatalf("Failed to receive stream response: %v", err)
		}
		if resp.Zone == protocol.SomaticZone_ZONE_UNSPECIFIED {
			t.Errorf("Received unspecified zone")
		}
		t.Logf("Received update %d: Zone=%v, Pressure=%d", i, resp.Zone, resp.PressureScore)
	}
}
