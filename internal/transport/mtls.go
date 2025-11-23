package transport

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// NewMTLSConfig creates a tls.Config configured for mTLS using the provided
// server certificate, key and CA certificate paths.
func NewMTLSConfig(serverCertPath, serverKeyPath, caCertPath string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(serverCertPath, serverKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load server cert/key: %w", err)
	}

	caCertPEM, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("read ca cert: %w", err)
	}

	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(caCertPEM); !ok {
		return nil, fmt.Errorf("failed to append CA cert")
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
	}

	return cfg, nil
}
