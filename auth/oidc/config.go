package oidc

import (
	"github.com/target/goalert/auth/nonce"
	"github.com/target/goalert/keyring"
)

// Config provides necessary parameters for OIDC authentication.
type Config struct {
	Keyring    keyring.Keyring
	NonceStore nonce.Store
}
