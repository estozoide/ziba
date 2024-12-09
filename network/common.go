package network

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// Server ports.

var (
	setupPort      = 9090
	accgenPort     = 9091
	withdrawalPort = 9092
	paymentPort    = 9093
	depositPort    = 9094
	exchangePort   = 9095
	getPort        = 9096
)

// CreateCertificate.
func CreateCertificate(baseDir string, baseName string) error {
	// Generate private key.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("failed to create private key: %v", err)
		return err
	}

	// Use certificate template.
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Carlos H. P."},
			CommonName:   "CHP",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
	}

	// Create certificate.
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatalf("failed to create certificate: %v", err)
		return err
	}

	// Save certificate to file.
	certFilename := fmt.Sprintf("%s_cert.pem", baseName)
	certPath := filepath.Join(baseDir, certFilename)
	certFile, err := os.Create(certPath)
	if err != nil {
		log.Fatalf("failed to create cert.pem: %v", err)
		return err
	}
	defer certFile.Close()

	// Encode DER bytes.
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		log.Fatalf("failed to encode certificate: %v", err)
		return err
	}

	// Save private key to file.
	keyFilename := fmt.Sprintf("%s_key.pem", baseName)
	keyPath := filepath.Join(baseDir, keyFilename)
	keyFile, err := os.Create(keyPath)
	if err != nil {
		log.Fatalf("failed to create key.pem")
		return err
	}
	defer keyFile.Close()

	// Read private key as DER bytes.
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Fatalf("failed to marshal private key: %v", err)
		return err
	}

	// Encode DER bytes.
	err = pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes})
	if err != nil {
		log.Fatalf("failed to encode private key bytes: %v", err)
		return err
	}

	return nil
}

// GetServerTLSConfig.
func GetServerTLSConfig(certPath, keyPath string) (*tls.Config, error) {
	// Load certificate and private key.
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Fatalf("failed to load certificate: %v", err)
		return nil, err
	}

	// Set TLS configuration.
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}

	return config, nil
}

// GetClientTLSConfig.
func GetClientTLSConfig(certPath string) (*tls.Config, error) {
	// Load certificate.
	cert, err := os.ReadFile(certPath)
	if err != nil {
		log.Fatalf("failed to read certificate: %v", err)
		return nil, err
	}

	// Create client's certificate pool.
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(cert) {
		log.Fatalf("failed to append cert to pool: %v", err)
		return nil, err
	}

	// Set TLS configuration.
	config := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
		ServerName: "localhost",
	}

	return config, nil
}
