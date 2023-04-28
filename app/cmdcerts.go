package app

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
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
	"golang.org/x/crypto/ed25519"
)

type certType int

const (
	certTypeUnknown certType = iota
	certTypeCASystem
	certTypeCAPlugin
	certTypeServer
	certTypeClient
)

func copyFile(dst, src string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read '%s': %w", src, err)
	}
	err = os.WriteFile(dst, data, 0o644)
	if err != nil {
		return fmt.Errorf("write '%s': %w", dst, err)
	}
	return nil
}

func loadPair(certFile, keyFile string) (cert *x509.Certificate, pk interface{}, err error) {
	data, err := os.ReadFile(certFile)
	if err != nil {
		return nil, nil, fmt.Errorf("read cert file '%s': %w", certFile, err)
	}
	p, _ := pem.Decode(data)
	cert, err = x509.ParseCertificate(p.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse cert file '%s': %w", certFile, err)
	}

	data, err = os.ReadFile(keyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("read key file '%s': %w", keyFile, err)
	}
	p, _ = pem.Decode(data)
	pk, err = x509.ParsePKCS8PrivateKey(p.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse key file '%s': %w", keyFile, err)
	}
	return cert, pk, nil
}

func certTemplate(t certType) *x509.Certificate {
	switch t {
	case certTypeCASystem, certTypeCAPlugin:
		return &x509.Certificate{
			IsCA:                  true,
			NotBefore:             time.Now(),
			NotAfter:              time.Now().AddDate(100, 0, 0),
			KeyUsage:              x509.KeyUsageCertSign,
			BasicConstraintsValid: true,
		}
	case certTypeServer, certTypeClient:
		return &x509.Certificate{
			Subject: pkix.Name{
				CommonName: _certCommonName, // Will be checked by the server
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().AddDate(100, 0, 0),
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
			DNSNames:              []string{_certCommonName},
		}
	}

	panic("unknown certType")
}

type keypair interface {
	Public() crypto.PublicKey
}

func privateKey() (keypair, error) {
	if _certED25519Key {
		_, pk, err := ed25519.GenerateKey(rand.Reader)
		return pk, err
	}

	switch _certECDSACurve {
	case "":
		// fall to RSA
	case "P224":
		return ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		return ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		return nil, fmt.Errorf("invalid ECDSA curve '%s'", _certECDSACurve)
	}

	return rsa.GenerateKey(rand.Reader, _certRSABits)
}

func genCertFiles(t certType, extra ...certType) error {
	template := certTemplate(t)
	sn, err := certSerialNumber()
	if err != nil {
		return err
	}

	pk, err := privateKey()
	if err != nil {
		return fmt.Errorf("generate private key: %w", err)
	}
	template.SerialNumber = sn
	parentCert, parentKey := template, (interface{})(pk)

	var certFile, keyFile string
	switch t {
	case certTypeCASystem:
		certFile = _certSystemCACertFile
		keyFile = _certSystemCAKeyFile
	case certTypeCAPlugin:
		certFile = _certPluginCACertFile
		keyFile = _certPluginCAKeyFile
	case certTypeServer:
		certFile = _certServerCertFile
		keyFile = _certServerKeyFile
		parentCert, parentKey, err = loadPair(_certSystemCACertFile, _certSystemCAKeyFile)
		if err != nil {
			return fmt.Errorf("load keypair: %w", err)
		}
		err = copyFile(_certServerCAFile, _certPluginCACertFile)
		if err != nil {
			return fmt.Errorf("copy CA bundle: %w", err)
		}
	case certTypeClient:
		certFile = _certClientCertFile
		keyFile = _certClientKeyFile
		parentCert, parentKey, err = loadPair(_certPluginCACertFile, _certPluginCAKeyFile)
		if err != nil {
			return fmt.Errorf("load keypair: %w", err)
		}
		err = copyFile(_certClientCAFile, _certSystemCACertFile)
		if err != nil {
			return fmt.Errorf("copy CA bundle: %w", err)
		}
	default:
		panic("unknown certType")
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, parentCert, pk.Public(), parentKey)
	if err != nil {
		return fmt.Errorf("create certificate: %w", err)
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("open cert file '%s': %w", certFile, err)
	}
	defer certOut.Close()

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		return fmt.Errorf("encode certificate: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(pk)
	if err != nil {
		return fmt.Errorf("encode private key: %w", err)
	}

	keyOut, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("open key file '%s': %w", keyFile, err)
	}
	defer keyOut.Close()

	err = pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return fmt.Errorf("encode private key: %w", err)
	}

	if len(extra) > 0 {
		return genCertFiles(extra[0], extra[1:]...)
	}
	return nil
}

var (
	genCerts = &cobra.Command{
		Use:   "gen-cert",
		Short: "Generate a certificate for SysAPI (gRPC) usage.",
	}

	genAllCert = &cobra.Command{
		Use:   "all",
		Short: "Generate all certificates for GoAlert to authenticate to/from gRPC clients.",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := genCertFiles(certTypeCASystem, certTypeCAPlugin, certTypeServer, certTypeClient)
			if err != nil {
				return fmt.Errorf("generate cert files: %w", err)
			}
			return nil
		},
	}

	genCACert = &cobra.Command{
		Use:   "ca",
		Short: "Generate a CA certificates for GoAlert to authenticate to/from gRPC clients.",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := genCertFiles(certTypeCASystem, certTypeCAPlugin)
			if err != nil {
				return fmt.Errorf("generate cert files: %w", err)
			}
			return nil
		},
	}

	genServerCert = &cobra.Command{
		Use:   "server",
		Short: "Generate a server certificate for GoAlert to authenticate to/from gRPC clients.",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := genCertFiles(certTypeServer)
			if err != nil {
				return fmt.Errorf("generate cert files: %w", err)
			}
			return nil
		},
	}
	genClientCert = &cobra.Command{
		Use:   "client",
		Short: "Generate a client certificate for services that talk to GoAlert.",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := genCertFiles(certTypeClient)
			if err != nil {
				return fmt.Errorf("generate cert files: %w", err)
			}
			return nil
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

var (
	_certCommonName       string = "GoAlert"
	_certSerialNumber     string = ""
	_certSystemCACertFile string = "system.ca.pem"
	_certSystemCAKeyFile  string = "system.ca.key"
	_certPluginCACertFile string = "plugin.ca.pem"
	_certPluginCAKeyFile  string = "plugin.ca.key"
	_certClientCertFile   string = "goalert-client.pem"
	_certClientKeyFile    string = "goalert-client.key"
	_certClientCAFile     string = "goalert-client.ca.pem"
	_certServerCertFile   string = "goalert-server.pem"
	_certServerKeyFile    string = "goalert-server.key"
	_certServerCAFile     string = "goalert-server.ca.pem"

	_certValidFrom  string        = ""
	_certValidFor   time.Duration = 10 * 365 * 24 * time.Hour
	_certRSABits    int           = 2048
	_certECDSACurve string        = ""
	_certED25519Key bool          = false
)

func initCertCommands() {
	genCerts.PersistentFlags().StringVar(&_certSerialNumber, "serial-number", _certSerialNumber, "Serial number to use for generated certificate (default is random).")

	genCerts.PersistentFlags().StringVar(&_certValidFrom, "start-date", _certValidFrom, "Creation date formatted as Jan 2 15:04:05 2006")
	genCerts.PersistentFlags().DurationVar(&_certValidFor, "duration", _certValidFor, "Creation date formatted as Jan 2 15:04:05 2006")
	genCerts.PersistentFlags().IntVar(&_certRSABits, "rsa-bits", _certRSABits, "Size of RSA key(s) to create. Ignored if either --ecdsa-curve or --ed25519 are set.")
	genCerts.PersistentFlags().StringVar(&_certECDSACurve, "ecdsa-curve", _certECDSACurve, "ECDSA curve to use to generate a key. Valid values are P224, P256 (recommended), P384, P521. Ignored if --ed25519 is set.")
	genCerts.PersistentFlags().BoolVar(&_certED25519Key, "ed25519", _certED25519Key, "Generate ED25519 key(s).")

	genCerts.PersistentFlags().StringVar(&_certCommonName, "cn", _certCommonName, "Common name of the certificate.")

	genCerts.PersistentFlags().StringVar(&_certSystemCACertFile, "system-ca-cert-file", _certSystemCACertFile, "CA cert file for signing server certs.")
	genCerts.PersistentFlags().StringVar(&_certSystemCAKeyFile, "system-ca-key-file", _certSystemCAKeyFile, "CA key file for signing server certs.")
	genCerts.PersistentFlags().StringVar(&_certPluginCACertFile, "plugin-ca-cert-file", _certPluginCACertFile, "CA cert file for signing client certs.")
	genCerts.PersistentFlags().StringVar(&_certPluginCAKeyFile, "plugin-ca-key-file", _certPluginCAKeyFile, "CA key file for signing client certs.")

	genServerCert.Flags().StringVar(&_certServerCertFile, "server-cert-file", _certServerCertFile, "Output file for the new server certificate.")
	genServerCert.Flags().StringVar(&_certServerKeyFile, "server-key-file", _certServerKeyFile, "Output file for the new server key.")
	genServerCert.Flags().StringVar(&_certServerCAFile, "server-ca-file", _certServerCAFile, "Output file for the server CA bundle.")

	genClientCert.Flags().StringVar(&_certClientCertFile, "client-cert-file", _certClientCertFile, "Output file for the new client certificate.")
	genClientCert.Flags().StringVar(&_certClientKeyFile, "client-key-file", _certClientKeyFile, "Output file for the new client key.")
	genClientCert.Flags().StringVar(&_certClientCAFile, "client-ca-file", _certClientCAFile, "Output file for the client CA bundle.")

	genCerts.AddCommand(genAllCert, genCACert, genServerCert, genClientCert)
}
