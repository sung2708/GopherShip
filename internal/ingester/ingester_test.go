package ingester

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/sungp/gophership/internal/buffer"
	"github.com/sungp/gophership/internal/stochastic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestNewIngester_BufferSizeValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"Valid Buffer", 512, 512},
		{"Zero Buffer (Default)", 0, 1024},
		{"Negative Buffer (Default)", -1, 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ing := NewIngester(tt.input)
			if cap(ing.buffer) != tt.expected {
				t.Errorf("NewIngester(%d) buffer capacity = %d; want %d", tt.input, cap(ing.buffer), tt.expected)
			}
		})
	}
}

func TestIngester_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ing := NewIngester(10)

	closed := make(chan struct{})
	go func() {
		ing.StartWorkerLoop(ctx)
		close(closed)
	}()

	cancel() // Worker should exit

	select {
	case <-closed:
		// Success: worker loop returned
	case <-time.After(1 * time.Second):
		t.Fatal("worker loop did not shut down on context cancellation")
	}
}

func TestIngester_SomaticPivotsCounter(t *testing.T) {
	ing := NewIngester(1) // Tiny buffer
	ctx := context.Background()

	// Fill buffer
	buf1 := buffer.MustAcquire(10)
	ing.IngestData(ctx, buf1)

	// Trigger pivot
	buf2 := buffer.MustAcquire(10)

	initialCount := testutil.ToFloat64(stochastic.SomaticPivotsTotal)

	ing.IngestData(ctx, buf2)

	finalCount := testutil.ToFloat64(stochastic.SomaticPivotsTotal)
	if finalCount != initialCount+1 {
		t.Errorf("expected SomaticPivotsTotal to increment by 1, got from %f to %f", initialCount, finalCount)
	}
}

// TestIngester_TLSVersionEnforcement verifies AC1 (Reject TLS < 1.3)
func TestIngester_TLSVersionEnforcement(t *testing.T) {
	// 1. Generate self-signed cert for testing
	cert, _ := generateTestCert(t)

	ing := NewIngester(10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}
	creds := credentials.NewTLS(tlsConfig)

	actualAddr, stop, err := ing.StartGRPCServer(ctx, "127.0.0.1:0", grpc.Creds(creds))
	if err != nil {
		t.Fatalf("failed to start gRPC server: %v", err)
	}
	defer stop()
	time.Sleep(100 * time.Millisecond) // Give server time to start

	// 2. Attempt connection with TLS 1.2 (Should fail)
	t.Run("Reject TLS 1.2", func(t *testing.T) {
		clientTLSConfig := &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS12,
		}
		conn, err := grpc.Dial(actualAddr, grpc.WithTransportCredentials(credentials.NewTLS(clientTLSConfig)))
		if err != nil {
			// This might not error immediately until we actually call something
			t.Logf("Dial returned error (expected in some gRPC versions): %v", err)
		} else {
			defer conn.Close()
			// Try a call - should fail during handshake
			// For simplicity in this mock, we'll just check if we can establish a raw TLS connection
			rawConn, err := tls.Dial("tcp", actualAddr, clientTLSConfig)
			if err == nil {
				rawConn.Close()
				t.Errorf("server accepted TLS 1.2 connection, want rejection")
			} else {
				t.Logf("TLS 1.2 connection correctly rejected: %v", err)
			}
		}
	})

	// 3. Attempt connection with TLS 1.3 (Should succeed)
	t.Run("Accept TLS 1.3", func(t *testing.T) {
		clientTLSConfig := &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS13,
		}
		rawConn, err := tls.Dial("tcp", actualAddr, clientTLSConfig)
		if err != nil {
			t.Errorf("server rejected TLS 1.3 connection: %v", err)
		} else {
			rawConn.Close()
			t.Log("TLS 1.3 connection accepted")
		}
	})
}

func generateTestCert(t *testing.T) (tls.Certificate, []byte) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"GopherShip Test"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}

	return cert, certPEM
}
