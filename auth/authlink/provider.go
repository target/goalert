package authlink

import (
	"context"
	"net/http"

	"github.com/target/goalert/auth"
	"github.com/target/goalert/config"
)

var _ auth.IdentityProvider = &Store{}

// Info implements the auth.Provider interface.
func (Store) Info(ctx context.Context) auth.ProviderInfo {
	cfg := config.FromContext(ctx)
	return auth.ProviderInfo{
		Title:   "Link Device",
		Hidden:  true,
		Enabled: !cfg.Auth.DisableBasic,
	}
}

// ExtractIdentity implements the auth.IdentityProvider interface handling both claim and auth token redemption.
func (s *Store) ExtractIdentity(route *auth.RouteInfo, w http.ResponseWriter, req *http.Request) (*auth.Identity, error) {

	switch route.RelativePath {
	case "/":
		// claim code

		return nil, nil
	case "/auth":
		// handled below
	default:
		return nil, auth.Error("Invalid callback URL specified in GitHub application config.")
	}

	return nil, nil
}
