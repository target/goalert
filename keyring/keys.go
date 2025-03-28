package keyring

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"

	"github.com/pkg/errors"
)

// Keys represents a set of encryption/decryption keys.
type Keys [][]byte

// Encrypt will encrypt and then encode data into PEM-format.
func (k Keys) Encrypt(label string, data []byte) ([]byte, error) {
	if len(k) == 0 {
		k = Keys{[]byte{}}
	}
	//nolint:all SA1019 TODO migrate off deprecated method; usage is secure for at-rest data
	block, err := x509.EncryptPEMBlock(rand.Reader, label, data, k[0], x509.PEMCipherAES256)
	if err != nil {
		return nil, err
	}
	data = pem.EncodeToMemory(block)
	return data, err
}

// Decrypt will decrypt PEM-encoded data using the first successful key. The index of
// the used key is returned as n.
func (k Keys) Decrypt(pemData []byte) (data []byte, label string, err error) {
	if len(k) == 0 {
		k = Keys{[]byte{}}
	}
	block, _ := pem.Decode(pemData)

	for _, key := range k {
		//nolint:all SA1019 TODO migrate off deprecated method; usage is secure for at-rest data
		data, err = x509.DecryptPEMBlock(block, key)
		if err == nil {
			return data, block.Type, nil
		}
	}

	return nil, "", errors.New("invalid decryption key")
}
