package stochastic

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestMetricsServer(t *testing.T) {
	ctx := context.Background()
	addr := "localhost:39091" // Use a different port for testing
	shutdown := StartMetricsServer(ctx, addr)
	defer shutdown()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://" + addr + "/metrics")
	if err != nil {
		t.Fatalf("failed to GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	content := string(body)
	if !strings.Contains(content, "gophership_ingester_zone_index") {
		t.Errorf("missing metric gophership_ingester_zone_index in /metrics output")
	}
	if !strings.Contains(content, "gophership_somatic_pivots_total") {
		t.Errorf("missing metric gophership_somatic_pivots_total in /metrics output")
	}
}

func TestIngesterZoneGauge(t *testing.T) {
	ctx := context.Background()
	// Start server if not already started
	addr := "localhost:39092"
	shutdown := StartMetricsServer(ctx, addr)
	defer shutdown()

	time.Sleep(100 * time.Millisecond)

	MustSetAmbientStatus(StatusRed)

	resp, err := http.Get("http://" + addr + "/metrics")
	if err != nil {
		t.Fatalf("failed to GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	if !strings.Contains(content, "gophership_ingester_zone_index 2") {
		t.Errorf("expected gauge value 2 for StatusRed, got: %s", content)
	}

	MustSetAmbientStatus(StatusGreen)
	resp2, _ := http.Get("http://" + addr + "/metrics")
	body2, _ := io.ReadAll(resp2.Body)
	content2 := string(body2)

	if !strings.Contains(content2, "gophership_ingester_zone_index 0") {
		t.Errorf("expected gauge value 0 for StatusGreen, got: %s", content2)
	}
}
