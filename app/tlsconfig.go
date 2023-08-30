package app

import (
	"crypto/tls"
	"fmt"

	"github.com/spf13/viper"
)

type tlsFlagPrefix string

func (t tlsFlagPrefix) CertFile() string { return viper.GetString(string(t) + "tls-cert-file") }
func (t tlsFlagPrefix) KeyFile() string  { return viper.GetString(string(t) + "tls-key-file") }
func (t tlsFlagPrefix) CertData() string { return viper.GetString(string(t) + "tls-cert-data") }
func (t tlsFlagPrefix) KeyData() string  { return viper.GetString(string(t) + "tls-key-data") }
func (t tlsFlagPrefix) Listen() string   { return viper.GetString(string(t) + "listen-tls") }

func (t tlsFlagPrefix) HasFiles() bool {
	return t.CertFile() != "" || t.KeyFile() != ""
}
func (t tlsFlagPrefix) HasData() bool {
	return t.CertData() != "" || t.KeyData() != ""
}
func (t tlsFlagPrefix) HasAny() bool {
	return t.HasFiles() || t.HasData() || t.Listen() != ""
}

// getTLSConfig creates a static TLS config using supplied certificate values.
// Returns nil if no certificate values are set.
func getTLSConfig(t tlsFlagPrefix) (*tls.Config, error) {
	if !t.HasAny() {
		return nil, nil
	}

	var cert tls.Certificate
	var err error
	switch {
	case t.HasFiles() == t.HasData(): // both set or unset
		return nil, fmt.Errorf("invalid tls config: exactly one of --%stls-cert-file and --%stls-key-file OR --%stls-cert-data and --%stls-key-data must be specified", t, t, t, t)
	case t.HasFiles():
		cert, err = tls.LoadX509KeyPair(t.CertFile(), t.KeyFile())
		if err != nil {
			return nil, fmt.Errorf("load tls cert files: %w", err)
		}
	case t.HasData():
		cert, err = tls.X509KeyPair([]byte(t.CertData()), []byte(t.KeyData()))
		if err != nil {
			return nil, fmt.Errorf("parse tls cert: %w", err)
		}
	}

	return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
}
