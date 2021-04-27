package app

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type certType int

const (
	certTypeUnknown certType = iota
	certTypeCA
	certTypeSystem
	certTypePlugin
)

func loadCert(prefix string) (p *pem.Block, cert *x509.Certificate, pk interface{}, err error) {
	caCertBytes, err := os.ReadFile(prefix + ".pem")
	if err != nil {
		return nil, nil, nil, err
	}
	certPem, _ := pem.Decode(caCertBytes)
	caCert, err := x509.ParseCertificate(certPem.Bytes)
	if err != nil {
		return nil, nil, nil, err
	}

	caKeyBytes, err := os.ReadFile(prefix + ".key")
	if err != nil {
		return nil, nil, nil, err
	}
	keyPem, _ := pem.Decode(caKeyBytes)
	caKey, err := x509.ParsePKCS8PrivateKey(keyPem.Bytes)
	if err != nil {
		return nil, nil, nil, err
	}
	return certPem, caCert, caKey, nil
}

func genCertFiles(template *x509.Certificate, filePrefix string, t certType) error {
	sn, err := certSerialNumber()
	if err != nil {
		return err
	}

	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	template.SerialNumber = sn
	parentCert := template
	parentKey := (interface{})(pk)

	switch t {
	case certTypeCA:
	case certTypeSystem:
		_, parentCert, parentKey, err = loadCert("plugin.ca")
	case certTypePlugin:
		_, parentCert, parentKey, err = loadCert("system.ca")
	default:
		panic("unknown certType")
	}
	if err != nil {
		return err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, parentCert, pk.Public(), parentKey)
	if err != nil {
		return err
	}

	certOut, err := os.Create(filePrefix + ".pem")
	if err != nil {
		return err
	}
	defer certOut.Close()

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		return err
	}

	var p *pem.Block
	switch t {
	case certTypeCA:
	case certTypeSystem:
		p, _, _, err = loadCert("system.ca")
	case certTypePlugin:
		p, _, _, err = loadCert("system.ca")
	default:
		panic("unknown certType")
	}
	if err != nil {
		return err
	}
	err = pem.Encode(certOut, p)
	if err != nil {
		return err
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(pk)
	if err != nil {
		return err
	}

	keyOut, err := os.Create(filePrefix + ".key")
	if err != nil {
		return err
	}
	defer keyOut.Close()

	err = pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return err
	}
	return nil
}

var (
	_certCommonName     string
	_certSerialNumber   string
	_certCACertFile     string
	_certCAKeyFile      string
	_certClientCertFile string
	_certClientKeyFile  string
	_certServerCertFile string
	_certServerKeyFile  string

	genCerts = &cobra.Command{
		Use:   "gen-cert",
		Short: "Generate a certificate for SysAPI (gRPC) usage.",
	}

	genCACert = &cobra.Command{
		Use:   "ca",
		Short: "Generate a CA certificates for GoAlert to authenticate to/from gRPC clients.",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := genCertFiles(&x509.Certificate{
				IsCA:                  true,
				NotBefore:             time.Now(),
				NotAfter:              time.Now().AddDate(100, 0, 0),
				KeyUsage:              x509.KeyUsageCertSign,
				BasicConstraintsValid: true,
			}, "system.ca", certTypeCA)
			if err != nil {
				return err
			}

			err = genCertFiles(&x509.Certificate{
				IsCA:                  true,
				NotBefore:             time.Now(),
				NotAfter:              time.Now().AddDate(100, 0, 0),
				KeyUsage:              x509.KeyUsageCertSign,
				BasicConstraintsValid: true,
			}, "plugin.ca", certTypeCA)
			if err != nil {
				return err
			}

			return nil
		},
	}

	genServerCert = &cobra.Command{
		Use:   "server",
		Short: "Generate a server certificate for GoAlert to authenticate to/from gRPC clients.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return genCertFiles(&x509.Certificate{
				Subject: pkix.Name{
					CommonName: _certCommonName, // Will be checked by the server
				},
				NotBefore:             time.Now(),
				NotAfter:              time.Now().AddDate(100, 0, 0),
				KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
				ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
				BasicConstraintsValid: true,
				DNSNames:              []string{_certCommonName},
			}, "goalert-server", certTypeSystem)
		},
	}
	genClientCert = &cobra.Command{
		Use:   "client",
		Short: "Generate a client certificate for services that talk to GoAlert.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return genCertFiles(&x509.Certificate{
				Subject: pkix.Name{
					CommonName: _certCommonName, // Will be checked by the server
				},
				NotBefore:             time.Now(),
				NotAfter:              time.Now().AddDate(100, 0, 0),
				KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
				ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
				BasicConstraintsValid: true,
				DNSNames:              []string{_certCommonName},
			}, "goalert-client", certTypePlugin)
		},
	}
)

func certSerialNumber() (*big.Int, error) {
	if _certSerialNumber == "" {
		return randSerialNumber(), nil
	}

	sn := new(big.Int)
	sn, ok := sn.SetString(_certSerialNumber, 10)
	if !ok {
		return nil, fmt.Errorf("invalid value for serial number '%s'", _certSerialNumber)
	}
	return sn, nil
}
func randSerialNumber() *big.Int {
	maxSN := new(big.Int)
	// x509 serial number can be up to 20 bytes, so 160 bits -1 (sign)
	maxSN.Exp(big.NewInt(2), big.NewInt(159), nil).Sub(maxSN, big.NewInt(1))
	sn, err := rand.Int(rand.Reader, maxSN)
	if err != nil {
		panic(err)
	}
	return sn
}

func initCertCommands() {
	genCerts.PersistentFlags().StringVar(&_certSerialNumber, "serial-number", "", "Serial number to use for generated certificate (default is random).")
	genCerts.PersistentFlags().StringVar(&_certCommonName, "cn", "GoAlert", "Common name of the certificate.")
	genCerts.PersistentFlags().StringVar(&_certCACertFile, "ca", "ca.crt", "Server/CA certificate file name (can be GoAlert server cert).")
	genCerts.PersistentFlags().StringVar(&_certCAKeyFile, "key", "ca.key", "ServerCA key file name (can be GoAlert server key).")

	genServerCert.Flags().StringVar(&_certServerCertFile, "server-cert", "server.crt", "Server/CA certificate file name (can be GoAlert server cert).")
	genServerCert.Flags().StringVar(&_certServerKeyFile, "server-key", "server.key", "ServerCA key file name (can be GoAlert server key).")

	genClientCert.Flags().StringVar(&_certClientCertFile, "client-cert", "client.crt", "Server/CA certificate file name (can be GoAlert server cert).")
	genClientCert.Flags().StringVar(&_certClientKeyFile, "client-key", "client.key", "ServerCA key file name (can be GoAlert server key).")

	genCerts.AddCommand(genCACert, genServerCert, genClientCert)
}
