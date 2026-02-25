package control

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sungp/gophership/internal/somatic"
	"github.com/sungp/gophership/internal/stochastic"
	"github.com/sungp/gophership/pkg/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	DefaultPort       = "9092"
	DefaultSocketPath = "/tmp/gophership.sock"
)

// Server represents the secure management plane gRPC server.
type Server struct {
	protocol.ControlServiceServer
	grpcServer *grpc.Server
	port       string
	socketPath string
	tlsConfig  *tls.Config
	startTime  time.Time
	somatic    *somatic.Controller
}

// NewServer initializes the GopherShip control plane.
func NewServer(port, socketPath string, tlsConfig *tls.Config, somatic *somatic.Controller) *Server {
	if port == "" {
		port = DefaultPort
	}
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}
	return &Server{
		port:       port,
		socketPath: socketPath,
		tlsConfig:  tlsConfig,
		startTime:  time.Now(),
		somatic:    somatic,
	}
}

// Start launches the management interface on secure sockets.
func (s *Server) Start(ctx context.Context) error {
	var opts []grpc.ServerOption

	// 1. Configure mTLS if provided (AC2)
	if s.tlsConfig != nil {
		creds := credentials.NewTLS(s.tlsConfig)
		opts = append(opts, grpc.Creds(creds))
	}

	s.grpcServer = grpc.NewServer(opts...)
	protocol.RegisterControlServiceServer(s.grpcServer, s)

	// Listen on TCP (mTLS)
	tcpAddr := fmt.Sprintf(":%s", s.port)
	tcpLis, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on TCP %s: %w", tcpAddr, err)
	}

	// Listen on Unix Domain Socket (UDS)
	// On Windows, UDS is supported in recent versions, but SO_PEERCRED is Linux only.
	if runtime.GOOS != "windows" {
		// Clean up existing socket
		if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove existing socket: %w", err)
		}
	}

	udsLis, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on Unix socket %s: %w", s.socketPath, err)
	}

	// Set socket permissions (NFR.Sec2)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(s.socketPath, 0660); err != nil {
			return fmt.Errorf("failed to set socket permissions: %w", err)
		}
	}

	log.Info().
		Str("port", s.port).
		Str("socket", s.socketPath).
		Bool("mtls_enabled", s.tlsConfig != nil).
		Msg("Starting GopherShip Control Plane")

	go func() {
		if err := s.grpcServer.Serve(tcpLis); err != nil {
			log.Error().Err(err).Msg("TCP gRPC server failed")
		}
	}()

	go func() {
		// Wrap UDS listener with security checks
		if err := s.grpcServer.Serve(newSecureListener(udsLis)); err != nil {
			log.Error().Err(err).Msg("Unix gRPC server failed")
		}
	}()

	return nil
}

// Stop gracefully shuts down the management server.
func (s *Server) Stop(ctx context.Context) error {
	log.Info().Msg("Stopping GopherShip Control Plane")
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	return nil
}

// Ping implements protocol.ControlServiceServer.
func (s *Server) Ping(ctx context.Context, _ *emptypb.Empty) (*protocol.PingResponse, error) {
	return &protocol.PingResponse{
		Version:       "0.1.0-dev",
		UptimeSeconds: int64(time.Since(s.startTime).Seconds()),
	}, nil
}

func (s *Server) WatchSomaticStatus(req *protocol.WatchStatusRequest, stream protocol.ControlService_WatchSomaticStatusServer) error {
	interval := time.Duration(req.RefreshIntervalMs) * time.Millisecond
	if interval < 100*time.Millisecond {
		interval = 1 * time.Second // Default/Safe floor
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Debug().Dur("interval", interval).Msg("Somatic status stream started")

	// Pre-allocate response object to eliminate heap churn in the loop (NFR.P1)
	resp := &protocol.StatusResponse{}

	for {
		select {
		case <-stream.Context().Done():
			log.Debug().Msg("Somatic status stream closed by client")
			return stream.Context().Err()
		case <-ticker.C:
			// Fetch and populate pre-allocated response
			if err := s.populateStatus(resp); err != nil {
				return err
			}
			if err := stream.Send(resp); err != nil {
				return err
			}
		}
	}
}

// populateStatus is a zero-allocation helper to fill a StatusResponse.
func (s *Server) populateStatus(resp *protocol.StatusResponse) error {
	status := stochastic.GetAmbientStatus()

	zone := protocol.SomaticZone_ZONE_GREEN
	switch status {
	case stochastic.StatusYellow:
		zone = protocol.SomaticZone_ZONE_YELLOW
	case stochastic.StatusRed:
		zone = protocol.SomaticZone_ZONE_RED
	}

	var usage, heap uint64
	var score uint32
	if stochastic.Monitor != nil {
		usage, heap, score = stochastic.Monitor.Telemetry()
	}

	resp.Zone = zone
	resp.PressureScore = score
	resp.MemoryUsageBytes = usage
	resp.HeapObjects = heap
	resp.GoroutineCount = uint32(runtime.NumGoroutine())

	return nil
}

// GetSomaticStatus implements protocol.ControlServiceServer.
func (s *Server) GetSomaticStatus(ctx context.Context, _ *emptypb.Empty) (*protocol.StatusResponse, error) {
	resp := &protocol.StatusResponse{}
	if err := s.populateStatus(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// OverrideSomaticZone implements protocol.ControlServiceServer.
func (s *Server) OverrideSomaticZone(ctx context.Context, req *protocol.OverrideSomaticZoneRequest) (*emptypb.Empty, error) {
	if s.somatic == nil {
		return nil, fmt.Errorf("somatic controller not initialized")
	}

	if req.Zone == protocol.SomaticZone_ZONE_UNSPECIFIED {
		s.somatic.ClearOverride()
		log.Info().Msg("Somatic zone override cleared (Manual Control End)")
		return &emptypb.Empty{}, nil
	}

	var target stochastic.AmbientStatus
	switch req.Zone {
	case protocol.SomaticZone_ZONE_GREEN:
		target = stochastic.StatusGreen
	case protocol.SomaticZone_ZONE_YELLOW:
		target = stochastic.StatusYellow
	case protocol.SomaticZone_ZONE_RED:
		target = stochastic.StatusRed
	default:
		return nil, fmt.Errorf("invalid somatic zone: %v", req.Zone)
	}

	s.somatic.Override(target)
	log.Info().
		Str("target_zone", target.String()).
		Msg("Somatic zone override triggered (Emergency Manual Override)")

	return &emptypb.Empty{}, nil
}

// LoadTLSConfig helper for loading mTLS credentials.
func LoadTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
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
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}, nil
}

// secureListener wraps a net.Listener to enforce SO_PEERCRED on Unix sockets.
type secureListener struct {
	net.Listener
}

func newSecureListener(l net.Listener) net.Listener {
	return &secureListener{l}
}

func (l *secureListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	// We only care about Unix Domain Sockets for SO_PEERCRED
	if conn.RemoteAddr().Network() == "unix" {
		if err := verifyPeer(conn); err != nil {
			log.Warn().Err(err).Str("addr", conn.RemoteAddr().String()).Msg("Unauthorized UDS connection rejected")
			conn.Close()
			return nil, fmt.Errorf("security violation: %w", err)
		}
	}

	return conn, nil
}
