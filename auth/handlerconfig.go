package auth

import (
	"github.com/target/goalert/apikey"
	"github.com/target/goalert/calsub"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/user"
)

// HandlerConfig provides configuration for the auth handler.
type HandlerConfig struct {
	UserStore      *user.Store
	SessionKeyring keyring.Keyring
	APIKeyring     keyring.Keyring
	IntKeyStore    *integrationkey.Store
	CalSubStore    *calsub.Store
	APIKeyStore    *apikey.Store
}
