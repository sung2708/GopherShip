package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gen_certs <output_dir>")
		os.Exit(1)
	}
	dir := os.Args[1]
	os.MkdirAll(dir, 0755)

	// 1. Generate CA
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"GopherShip Test CA"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		panic(err)
	}

	savePEM(dir+"/ca.crt", "CERTIFICATE", caBytes)
	savePEM(dir+"/ca.key", "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(caKey))

	// 2. Generate Server Cert
	genCert(dir, "server", caTemplate, caKey)

	// 3. Generate Client Cert
	genCert(dir, "client", caTemplate, caKey)

	fmt.Println("Certificates generated in", dir)
}

func genCert(dir, name string, caTemplate *x509.Certificate, caKey *rsa.PrivateKey) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: name,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour * 24),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}

	certBytes, _ := x509.CreateCertificate(rand.Reader, template, caTemplate, &key.PublicKey, caKey)
	savePEM(dir+"/"+name+".crt", "CERTIFICATE", certBytes)
	savePEM(dir+"/"+name+".key", "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))
}

func savePEM(filename, typeName string, data []byte) {
	f, _ := os.Create(filename)
	defer f.Close()
	pem.Encode(f, &pem.Block{Type: typeName, Bytes: data})
}
