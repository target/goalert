package authlink

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/target/goalert/auth"
	"github.com/target/goalert/config"
)

var _ auth.IdentityProvider = &Store{}

// Info implements the auth.Provider interface.
func (Store) Info(ctx context.Context) auth.ProviderInfo {
	cfg := config.FromContext(ctx)
	return auth.ProviderInfo{
		Title:      "Link Device",
		Hidden:     true,
		Enabled:    !cfg.Auth.DisableAuthLink,
		NoRedirect: true,
	}
}

// ExtractIdentity implements the auth.IdentityProvider interface handling both claim and auth token redemption.
func (s *Store) ExtractIdentity(route *auth.RouteInfo, w http.ResponseWriter, req *http.Request) (*auth.Identity, error) {
	switch route.RelativePath {
	case "/":
		resp, err := s.Claim(req.Context(), req.FormValue("code"), true)
		if err == ErrBadID {
			return nil, auth.Error("invalid code")
		}
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(resp)
		if err != nil {
			return nil, err
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return nil, auth.ErrAlreadyResponded
	case "/auth":
		// handled below
	default:
		return nil, auth.Error("invalid route")
	}

	userID, err := s.Auth(req.Context(), req.FormValue("AuthToken"))
	if err != nil {
		return nil, err
	}

	return &auth.Identity{UserID: userID}, nil
}
