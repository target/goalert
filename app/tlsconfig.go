package app

import (
	"crypto/tls"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// getTLSConfig creates a static TLS config using supplied certificate values.
// Returns nil if no certificate values are set.
func getTLSConfig(certFileFlag string, keyFileFlag string, certDataFlag string, keyDataFlag string, listenFlag string) (*tls.Config, error) {

	var n int
	if viper.GetString(certFileFlag) != "" {
		n += 0b0001
	}
	if viper.GetString(keyFileFlag) != "" {
		n += 0b0010
	}
	if viper.GetString(certDataFlag) != "" {
		n += 0b0100
	}
	if viper.GetString(keyDataFlag) != "" {
		n += 0b1000
	}

	var cert tls.Certificate
	var err error
	switch n {
	case 0b0011: // file mode
		cert, err = tls.LoadX509KeyPair(viper.GetString(certFileFlag), viper.GetString(keyFileFlag))
		if err != nil {
			return nil, errors.Wrap(err, "load tls cert files")
		}
	case 0b1100: // data mode
		cert, err = tls.X509KeyPair([]byte(viper.GetString(certDataFlag)), []byte(viper.GetString(keyDataFlag)))
		if err != nil {
			return nil, errors.Wrap(err, "parse tls cert")
		}
	case 0: // no flags set
		if viper.GetString(listenFlag) == "" {
			return nil, nil
		}
		fallthrough
	default:
		return nil, errors.New("--tls-cert-file and --tls-key-file OR --tls-cert-data and --tls-key-data must be specified")
	}

	return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
}
