package stochastic

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var (
	// Registry is the custom Prometheus registry for GopherShip metrics.
	Registry = prometheus.NewRegistry()

	// IngesterZone tracks the current somatic zone.
	// 0: Green, 1: Yellow, 2: Red.
	IngesterZone = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gophership_ingester_zone_index",
		Help: "Current somatic zone index (0: Green, 1: Yellow, 2: Red).",
	})

	// SomaticPivotsTotal tracks the total number of fallback triggers.
	SomaticPivotsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gophership_somatic_pivots_total",
		Help: "Total number of somatic pivots (reflex triggers).",
	})

	// IngesterUsageBytes tracks active memory usage of the ingester.
	IngesterUsageBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gophership_ingester_usage_bytes",
		Help: "Active memory usage of the Ingester in bytes.",
	})

	// VaultUsageBytes tracks active memory usage of the vault.
	VaultUsageBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gophership_vault_usage_bytes",
		Help: "Active memory usage of the Vault (WALsegments + blocks) in bytes.",
	})
)

func init() {
	// Register metrics with the custom registry.
	Registry.MustRegister(IngesterZone)
	Registry.MustRegister(SomaticPivotsTotal)
	Registry.MustRegister(IngesterUsageBytes)
	Registry.MustRegister(VaultUsageBytes)

	// Initialize the zone to Green (0).
	IngesterZone.Set(0)
}

// StartMetricsServer starts a Prometheus metrics server on the specified address.
// It returns a cleanup function that should be called during graceful shutdown.
func StartMetricsServer(ctx context.Context, addr string) func() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(Registry, promhttp.HandlerOpts{}))

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Info().Str("addr", addr).Msg("Starting Prometheus metrics server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Prometheus metrics server failed")
		}
	}()

	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		log.Info().Msg("Shutting down Prometheus metrics server...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("Prometheus metrics server shutdown failed")
		}
	}
}
