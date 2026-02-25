package otel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

// Telemetry manages the OpenTelemetry provider lifecycle.
type Telemetry struct {
	// TODO: Add tracer and meter provider handles
}

// CreateIngestionTLSConfig prepares a secure TLS 1.3 configuration.
func CreateIngestionTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load ingestion key pair: %w", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		// Note: CipherSuites are automatically managed for TLS 1.3
	}

	if caFile != "" {
		caCert, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %w", err)
		}
		caPool := x509.NewCertPool()
		if ok := caPool.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("failed to append CA cert to pool (no certs found in PEM)")
		}
		config.ClientCAs = caPool
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return config, nil
}

// Setup initializes the OTLP exporters and global providers.
func Setup(ctx context.Context, serviceName string) (*Telemetry, error) {
	log.Info().Str("service", serviceName).Msg("Initializing OpenTelemetry baseline")
	// TODO: Configure OTLP/gRPC exporter targets
	return &Telemetry{}, nil
}

// Shutdown ensures all telemetry spans and metrics are flushed.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	log.Info().Msg("Shutting down OpenTelemetry providers")
	return nil
}
