package main

import (
	"context"
	"crypto/tls"
	"os/signal"
	"syscall"
	"time"

	"github.com/sungp/gophership"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sungp/gophership/internal/config"
	"github.com/sungp/gophership/internal/control"
	"github.com/sungp/gophership/internal/ingester"
	"github.com/sungp/gophership/internal/stochastic"
	"github.com/sungp/gophership/internal/web"
	"github.com/sungp/gophership/pkg/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	Version = "1.0.0"
	Commit  = "none"
)

func main() {
	// Initialize structured logging (Architecture Mandate)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().
		Str("version", Version).
		Str("commit", Commit).
		Msg("Starting GopherShip Engine (Tier 1 Foundation)")

	// Setup Signal-Linked Context for Graceful Shutdown (NFR.DP1)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Load Configuration (Story 6.2)
	cfg, err := config.Load("")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to load configuration; using defaults")
		cfg = config.Default()
	}

	// === Phase Initiation ===
	// 0. Initialize Global Stochastic Monitor (Story 4.1 & 4.2)
	// Budgets: Centralized in internal/config
	monitor := stochastic.NewSensingMonitor(
		1024,                  // Check every 1024 ops
		cfg.Monitoring.MaxRAM, // Max Host RAM
		cfg.Monitoring.YellowThreshold,
		cfg.Monitoring.RedThreshold,   // Thresholds
		cfg.Monitoring.IngesterBudget, // Ingester Budget
		cfg.Monitoring.VaultBudget,    // Vault Budget
	)
	stochastic.SetGlobalMonitor(monitor)

	// 1. Initialize Ingester (Core Reflex Engine)
	ing := ingester.NewIngester(cfg.Ingester.BufferSize)
	ing.StartWorkerLoop(ctx)

	// 2. Start OTLP gRPC Ingestion Server (GS.1.2)
	// [AC1, AC3, AC4] TLS 1.3 and mTLS Configuration
	var grpcOpts []grpc.ServerOption

	certFile := cfg.Ingester.TLS.CertFile
	keyFile := cfg.Ingester.TLS.KeyFile
	caFile := cfg.Ingester.TLS.CAFile

	if certFile != "" || keyFile != "" {
		if certFile == "" || keyFile == "" {
			log.Fatal().Msg("Incomplete TLS configuration: both GS_INGEST_CERT and GS_INGEST_KEY (or YAML equivalents) must be provided")
		}
		tlsConfig, err := otel.CreateIngestionTLSConfig(certFile, keyFile, caFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create TLS config")
		}
		grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewTLS(tlsConfig)))
		if caFile != "" {
			log.Info().Str("ca", caFile).Msg("mTLS enabled for ingestion (CA pool initialized)")
		} else {
			log.Info().Msg("TLS 1.3 ingestion enabled")
		}
	} else {
		log.Warn().Msg("Ingestion server running WITHOUT TLS (Insecure)")
	}

	_, stopGRPC, err := ing.StartGRPCServer(ctx, cfg.Ingester.Addr, grpcOpts...)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start OTLP gRPC server")
	}

	// 3. Start Prometheus Metrics Server (AC5)
	metricsShutdown := stochastic.StartMetricsServer(ctx, ":9091")
	defer metricsShutdown()

	// 4. Initialize Secure Control Plane (Story 5.1)
	var ctrlTLS *tls.Config
	if certFile != "" && keyFile != "" && caFile != "" {
		c, err := control.LoadTLSConfig(certFile, keyFile, caFile)
		if err != nil {
			log.Error().Err(err).Msg("Failed to load control plane TLS config")
		} else {
			ctrlTLS = c
		}
	}

	ctrl := control.NewServer("", "./gophership.sock", ctrlTLS, ing.Somatic())
	if err := ctrl.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to start control plane")
	}
	defer ctrl.Stop(ctx)

	// 5. Start Web UI Server (Story 6.2 integration)
	webSrv := web.NewMetricsServer(gophership.DashboardAssets)
	if err := webSrv.Start(ctx, ":8080"); err != nil {
		log.Fatal().Err(err).Msg("Failed to start Web UI server")
	}

	// Keep alive until signal (NFR.DP1)
	log.Info().Msg("GopherShip Engine is now active and sensitizing...")
	<-ctx.Done()

	// === Graceful Shutdown (NFR.DP1) ===
	// All Tier 1 components (Ingester, Vault) respond to ctx.Done().
	// Wait for gRPC server to stop (timed cutoff for NFR.DP1)
	log.Info().Msg("Stopping OTLP gRPC Server...")

	stopDone := make(chan struct{})
	go func() {
		stopGRPC()
		close(stopDone)
	}()

	select {
	case <-stopDone:
		log.Info().Msg("gRPC server stopped gracefully")
	case <-time.After(5 * time.Second):
		log.Warn().Msg("gRPC graceful shutdown timed out; forcing stop")
	}

	log.Info().Msg("Shutdown signal received. Initiating graceful shutdown sequence...")
	log.Info().Msg("Shutdown sequence complete. All biological reflexes stopped.")
}
