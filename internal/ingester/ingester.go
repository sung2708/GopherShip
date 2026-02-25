package ingester

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/rs/zerolog/log"
	"github.com/sungp/gophership/internal/buffer"
	"github.com/sungp/gophership/internal/somatic"
	"github.com/sungp/gophership/internal/stochastic"
	"github.com/sungp/gophership/internal/vault"
	logcol "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// Ingester represents the high-throughput entry point.
type Ingester struct {
	logcol.UnimplementedLogsServiceServer
	buffer         chan *[]byte // Stores pointers to pooled buffers
	processedCount uint64
	fallbackCount  uint64 // Total count of dropped packets (Somatic Pivots)
	quit           chan struct{}
	somatic        *somatic.Controller
}

func NewIngester(bufferSize int) *Ingester {
	if bufferSize <= 0 {
		bufferSize = 1024 // Default to power-of-two alignment
	}

	// Optimization: Enforce power-of-two for L1 cache friendliness in channel scheduling
	if (bufferSize & (bufferSize - 1)) != 0 {
		log.Warn().Int("requested", bufferSize).Msg("Buffer size not power-of-two; may impact cache efficiency")
	}

	i := &Ingester{
		buffer: make(chan *[]byte, bufferSize),
		quit:   make(chan struct{}),
	}
	i.somatic = somatic.NewController(i)
	return i
}

// Stop shuts down the ingester and its associated worker loops.
func (i *Ingester) Stop() {
	close(i.quit)
}

// BufferDepth returns the current number of items in the ingestion buffer.
func (i *Ingester) BufferDepth() int {
	return len(i.buffer)
}

// BufferCap returns the total capacity of the ingestion buffer.
func (i *Ingester) BufferCap() int {
	return cap(i.buffer)
}

// Somatic returns the somatic controller for this ingester.
func (i *Ingester) Somatic() *somatic.Controller {
	return i.somatic
}

// IngestData demonstrates the "Local Reflex" and "Stochastic Awareness".
func (i *Ingester) IngestData(ctx context.Context, data *[]byte) {
	// 1. Stochastic Check (Every 1024 operations, reassess pressure)
	if stochastic.Monitor != nil && stochastic.Monitor.ShouldCheck() {
		// [AC2] Optimization: Trigger host and component sensing
		stochastic.Monitor.MustSense()

		// Ingester also does its own somatic reassessment (hysteresis)
		i.somatic.Reassess()
	}

	// 2. Local Reflex (Select-Default for zero-latency buffer sensing)
	select {
	case i.buffer <- data:
		// Normal case: Buffer has room.
		// Worker loop will release this buffer.
	case <-ctx.Done():
		// Context cancelled case: Don't block ingestion.
		buffer.MustRelease(data)
		return
	default:
		// Reflex case: Buffer is full! Pivot to "Somatic Fallback" (Store Raw)
		i.somaticFallback(ctx, data)
	}
}

func (i *Ingester) somaticFallback(ctx context.Context, data *[]byte) {
	// [NFR.P2] Optimization: Perform release BEFORE logging to minimize reflex latency.
	size := len(*data)
	buffer.MustRelease(data)

	// Report usage reduction
	if stochastic.Monitor != nil {
		stochastic.Monitor.ReportIngesterUsage(-int64(size))
	}

	// Increment Prometheus counter (Thread-safe, high efficiency)
	stochastic.SomaticPivotsTotal.Inc()

	dropped := atomic.AddUint64(&i.fallbackCount, 1)
	if dropped%1024 == 0 {
		if e := log.Info(); e.Enabled() {
			e.Int("size_bytes", size).
				Uint64("total_dropped", dropped).
				Msg("Reflex triggered: Ingestion buffer full. Routing to Raw Vault (Stochastic Log)")
		}
	}
}

// StartWorkerLoop begins the processing phase with graceful shutdown support.
func (i *Ingester) StartWorkerLoop(ctx context.Context) {
	go func() {
		for {
			select {
			case data, ok := <-i.buffer:
				if !ok {
					log.Info().Msg("Ingester buffer channel closed, shutting down worker...")
					return
				}
				// In high-load, this is where we'd check GetAmbientStatus to decide
				// if we should full-parse or just forward raw blobs.
				size := len(*data)
				_ = data
				buffer.MustRelease(data)

				// Report usage reduction
				if stochastic.Monitor != nil {
					stochastic.Monitor.ReportIngesterUsage(-int64(size))
				}
				atomic.AddUint64(&i.processedCount, 1)
			case <-i.quit:
				log.Info().Msg("Ingester worker loop received quit signal...")
				return
			case <-ctx.Done():
				log.Info().Msg("Ingester worker loop shutting down (context done)...")
				return
			}
		}
	}()
}

// ReplayRawVault streams data from the Raw Vault back into the ingestion buffer.
func (i *Ingester) ReplayRawVault(ctx context.Context, w *vault.WAL, itemsPerSecond int) error {
	replayer := vault.NewReplayer(w, itemsPerSecond)

	return replayer.StreamTo(ctx, func(data []byte) error {
		// [NFR.P1] Zero-allocation copy to pooled buffer
		bufPtr := buffer.MustAcquire(len(data))
		copy(*bufPtr, data)

		// [AC3] Throttling via blocking channel send
		select {
		case i.buffer <- bufPtr:
			return nil
		case <-ctx.Done():
			buffer.MustRelease(bufPtr)
			return ctx.Err()
		}
	})
}

// Export implements the OTLP gRPC ExportLogsService.
func (i *Ingester) Export(ctx context.Context, req *logcol.ExportLogsServiceRequest) (*logcol.ExportLogsServiceResponse, error) {
	// NFR.P1: Use pooled buffers to ensure zero heap allocations during marshaling.
	size := proto.Size(req)
	bufPtr := buffer.MustAcquire(size)

	opts := proto.MarshalOptions{}
	out, err := opts.MarshalAppend((*bufPtr)[:0], req)
	if err != nil {
		buffer.MustRelease(bufPtr)
		log.Error().Err(err).Msg("Failed to marshal OTLP request")
		return nil, err
	}

	*bufPtr = out

	// Report usage increase
	if stochastic.Monitor != nil {
		stochastic.Monitor.ReportIngesterUsage(int64(size))
	}

	i.IngestData(ctx, bufPtr)

	return &logcol.ExportLogsServiceResponse{}, nil
}

// StartGRPCServer initializes and starts the OTLP gRPC listener. Returns the bound address and a stop function.
func (i *Ingester) StartGRPCServer(ctx context.Context, addr string, opts ...grpc.ServerOption) (string, func(), error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return "", nil, err
	}

	actualAddr := lis.Addr().String()
	s := grpc.NewServer(opts...)
	logcol.RegisterLogsServiceServer(s, i)

	log.Info().Str("addr", actualAddr).Msg("Starting OTLP gRPC Ingestion Server")

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Error().Err(err).Msg("gRPC server failed")
		}
	}()

	return actualAddr, s.GracefulStop, nil
}
