package sysapi

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
)

// NewTLS will generate a new `*tls.Config` for use with a client or server.
func NewTLS(caFile, certFile, keyFile string) (*tls.Config, error) {
	tlsCfg := &tls.Config{ServerName: "GoAlert"}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsCfg.Certificates = append(tlsCfg.Certificates, cert)

	if caFile != "" {
		// If CA file is specified, require client auth
		caBytes, err := os.ReadFile(caFile)
		if err != nil {
			return nil, err
		}

		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caBytes) {
			return nil, errors.New("failed to append CA certs from PEM")
		}

		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
		tlsCfg.RootCAs = pool
		tlsCfg.ClientCAs = pool
	}

	return tlsCfg, nil
}
