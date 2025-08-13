package rabbitmqx

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// createTLSConfig creates a TLS configuration for RabbitMQ connection
func createTLSConfig(caCertFile, certFile, keyFile string, skipVerify bool) (*tls.Config, error) {
	// Create a TLS configuration with the certificate of the CA that signed the server's certificate
	rootCAs := x509.NewCertPool()

	if skipVerify {
		return &tls.Config{
			RootCAs:            rootCAs,
			InsecureSkipVerify: skipVerify,
		}, nil
	}

	if caCertFile != "" {
		caCert, err := os.ReadFile(caCertFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		if !rootCAs.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}
	}

	// If client certificates are provided, load them
	var certificates []tls.Certificate
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		certificates = append(certificates, cert)
	}

	return &tls.Config{
		RootCAs:      rootCAs,
		Certificates: certificates,
		MinVersion:   tls.VersionTLS12,
	}, nil
}
