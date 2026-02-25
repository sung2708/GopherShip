package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sungp/gophership/pkg/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	Version = "0.1.0-dev"
	Commit  = "none"
)

func main() {
	// CLI Flags
	socketPath := flag.String("socket", "/tmp/gophership.sock", "Path to the Unix Domain Socket")
	addr := flag.String("addr", "localhost:9092", "Address of the remote gRPC control plane (for mTLS)")
	certFile := flag.String("cert", "", "Path to the client certificate (for mTLS)")
	keyFile := flag.String("key", "", "Path to the client key (for mTLS)")
	caFile := flag.String("ca", "", "Path to the CA certificate (for mTLS)")
	useTLS := flag.Bool("tls", false, "Use mTLS instead of Unix Domain Socket")
	outputFormat := flag.String("output", "table", "Output format (table, json, yaml)")
	mockZone := flag.Int("mock-zone", -1, "Mock somatic zone for testing (0=Green, 1=Yellow, 2=Red)")
	overrideZone := flag.String("zone", "", "Somatic zone to force (green, yellow, red, none)")
	refreshInterval := flag.Duration("refresh", 1*time.Second, "Refresh interval for the dashboard (e.g. 500ms, 2s)")
	flag.Parse()

	// Initialize structured logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Setup Signal-Linked Context for Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Command Execution: status or top
	args := flag.Args()
	cmd := "status"
	if len(args) > 0 {
		cmd = args[0]
	}

	switch cmd {
	case "top":
		// Initialize Dashboard
		dash := NewDashboard()

		// Setup Connection
		opCtx, opCancel := context.WithCancel(ctx)
		defer opCancel()

		conn, err := dialControlPlane(opCtx, *addr, *socketPath, *useTLS, *certFile, *keyFile, *caFile)
		if err != nil {
			log.Error().Err(err).Msg("Failed to connect to control plane")
			os.Exit(129)
		}
		defer conn.Close()

		client := protocol.NewControlServiceClient(conn)

		// Start Streaming
		refreshMs := uint32((*refreshInterval).Milliseconds())
		stream, err := client.WatchSomaticStatus(opCtx, &protocol.WatchStatusRequest{
			RefreshIntervalMs: refreshMs,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to start somatic status stream")
			os.Exit(128)
		}

		// Background receiver
		go func() {
			for {
				resp, err := stream.Recv()
				if err != nil {
					log.Debug().Err(err).Msg("Stream closed")
					dash.Stop()
					return
				}
				dash.Update(resp)
			}
		}()

		// Run Dashboard (Blocks until exit)
		if err := dash.Run(); err != nil {
			log.Error().Err(err).Msg("Dashboard failure")
			os.Exit(1)
		}

	case "status":
		var resp *protocol.StatusResponse

		if *mockZone >= 0 {
			// Map intuitive values to proto enums (AC-Review)
			var zone protocol.SomaticZone
			switch *mockZone {
			case 0:
				zone = protocol.SomaticZone_ZONE_GREEN
			case 1:
				zone = protocol.SomaticZone_ZONE_YELLOW
			case 2:
				zone = protocol.SomaticZone_ZONE_RED
			default:
				zone = protocol.SomaticZone_ZONE_UNSPECIFIED
			}
			resp = &protocol.StatusResponse{
				Zone: zone,
			}
		} else {
			// Establish Secure gRPC Connection only when needed
			log.Debug().Msg("Initializing gRPC transport...")

			// We use a shared timeout for both dial and the target RPC to ensure 5s total budget
			opCtx, opCancel := context.WithTimeout(ctx, 5*time.Second)
			defer opCancel()

			conn, err := dialControlPlane(opCtx, *addr, *socketPath, *useTLS, *certFile, *keyFile, *caFile)
			if err != nil {
				log.Error().Err(err).Msg("Failed to initialize gRPC dialer")
				os.Exit(129)
			}
			defer conn.Close()

			log.Debug().Msg("Transport initialized (async), requesting somatic status...")
			client := protocol.NewControlServiceClient(conn)
			resp, err = client.GetSomaticStatus(opCtx, &emptypb.Empty{})
			if err != nil {
				diagnoseAndExit(err)
			}
		}

		exitCode := executeStatus(resp, *outputFormat)
		os.Exit(exitCode)

	case "override":
		zoneStr := strings.ToLower(*overrideZone)
		var target protocol.SomaticZone
		switch zoneStr {
		case "green":
			target = protocol.SomaticZone_ZONE_GREEN
		case "yellow":
			target = protocol.SomaticZone_ZONE_YELLOW
		case "red":
			target = protocol.SomaticZone_ZONE_RED
		case "none", "":
			target = protocol.SomaticZone_ZONE_UNSPECIFIED
		default:
			log.Error().Str("zone", zoneStr).Msg("Invalid zone. Must be green, yellow, red, or none")
			os.Exit(1)
		}

		opCtx, opCancel := context.WithTimeout(ctx, 5*time.Second)
		defer opCancel()

		conn, err := dialControlPlane(opCtx, *addr, *socketPath, *useTLS, *certFile, *keyFile, *caFile)
		if err != nil {
			log.Error().Err(err).Msg("Failed to connect to control plane")
			os.Exit(129)
		}
		defer conn.Close()

		client := protocol.NewControlServiceClient(conn)
		_, err = client.OverrideSomaticZone(opCtx, &protocol.OverrideSomaticZoneRequest{
			Zone: target,
		})
		if err != nil {
			diagnoseAndExit(err)
		}

		if target == protocol.SomaticZone_ZONE_UNSPECIFIED {
			fmt.Println("Somatic override cleared. Sensor control restored.")
		} else {
			fmt.Printf("Somatic zone successfully overridden to %s\n", strings.ToUpper(zoneStr))
		}
		os.Exit(0)

	default:
		fmt.Printf("GopherShip %s (%s)\n", Version, Commit)
		fmt.Println("Usage: gs-ctl [flags] <command>")
		fmt.Println("\nCommands:")
		fmt.Println("  status    Show internal engine health and pressure zone")
		fmt.Println("  top       Start a real-time somatic dashboard")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
	}
}

// dialControlPlane centralizes the gRPC connection logic for CLI commands.
func dialControlPlane(ctx context.Context, addr, socketPath string, useTLS bool, certFile, keyFile, caFile string) (*grpc.ClientConn, error) {
	if useTLS {
		log.Debug().Str("addr", addr).Msg("Configuring mTLS transport...")
		tlsConfig, err := loadClientTLSConfig(certFile, keyFile, caFile)
		if err != nil {
			return nil, fmt.Errorf("TLS config failed: %w", err)
		}
		creds := credentials.NewTLS(tlsConfig)
		return grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(creds))
	}

	log.Debug().Str("socket", socketPath).Msg("Configuring UDS transport...")
	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return net.Dial("unix", addr)
	}
	return grpc.DialContext(ctx, socketPath,
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func executeStatus(s *protocol.StatusResponse, format string) int {
	f, err := NewFormatter(format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 128
	}

	data := map[string]interface{}{
		"Somatic Zone":   s.Zone.String(),
		"Pressure Score": fmt.Sprintf("%d%%", s.PressureScore),
		"Memory Usage":   fmt.Sprintf("%d bytes", s.MemoryUsageBytes),
		"Heap Objects":   s.HeapObjects,
		"Goroutines":     s.GoroutineCount,
	}

	if format == "table" {
		fmt.Println("GopherShip Engine Status")
		fmt.Println("=========================")
	}

	if err := f.Format(os.Stdout, data); err != nil {
		log.Error().Err(err).Msg("Error formatting output")
		return 128
	}

	// AC4: Exit codes based on Somatic Zone
	switch s.Zone {
	case protocol.SomaticZone_ZONE_GREEN:
		return 0
	case protocol.SomaticZone_ZONE_YELLOW:
		return 1
	case protocol.SomaticZone_ZONE_RED:
		return 2
	default:
		return 0
	}
}

func loadClientTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	if certFile == "" || keyFile == "" || caFile == "" {
		return nil, fmt.Errorf("cert, key, and ca flags are required for mTLS")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	caPool := x509.NewCertPool()
	caData, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	if !caPool.AppendCertsFromPEM(caData) {
		return nil, fmt.Errorf("failed to append CA certs")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
	}, nil
}

func diagnoseAndExit(err error) {
	msg := err.Error()
	log.Debug().Err(err).Msg("Analyzing connection error")

	// 1. Connectivity Issues (Exit 129)
	if strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "context deadline exceeded") ||
		strings.Contains(msg, "actively refused it") ||
		strings.Contains(msg, "Unavailable") {
		log.Error().Msg("DIAGNOSTIC: Connection Refused or Timeout. Ensure GopherShip engine is running and address/socket is correct.")
		os.Exit(129)
	}

	// 2. Authentication/TLS Issues (Exit 130)
	if strings.Contains(msg, "tls: bad certificate") ||
		strings.Contains(msg, "x509: certificate signed by unknown authority") ||
		strings.Contains(msg, "remote error: tls") {
		log.Error().Msg("DIAGNOSTIC: mTLS Handshake Failed. Verify client certificates, keys, and CA trust.")
		os.Exit(130)
	}

	// 3. General gRPC/Application Errors (Exit 128)
	log.Error().Err(err).Msg("Failed to retrieve somatic status")
	os.Exit(128)
}
