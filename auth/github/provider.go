package github

import (
	"context"
	"strings"

	"github.com/target/goalert/config"
	"golang.org/x/oauth2"
	o2Github "golang.org/x/oauth2/github"
)

// Provider will respond to /auth and /callback endpoints for the purposes of GitHub OAuth2 authentication.
type Provider struct {
	c Config
}

func authConfig(ctx context.Context) *oauth2.Config {
	cfg := config.FromContext(ctx)

	authURL := o2Github.Endpoint.AuthURL
	tokenURL := o2Github.Endpoint.TokenURL
	if cfg.GitHub.EnterpriseURL != "" {
		authURL = strings.TrimSuffix(cfg.GitHub.EnterpriseURL, "/") + "/login/oauth/authorize"
		tokenURL = strings.TrimSuffix(cfg.GitHub.EnterpriseURL, "/") + "/login/oauth/access_token"
	}
	scopes := []string{"read:user"}
	if len(cfg.GitHub.AllowedOrgs) > 0 {
		scopes = append(scopes, "read:org")
	}
	return &oauth2.Config{
		ClientID:     cfg.GitHub.ClientID,
		ClientSecret: cfg.GitHub.ClientSecret,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}
}

// NewProvider will validate Config and create a new Provider. If Enabled is false, validation
// will be skipped.
func NewProvider(ctx context.Context, c *Config) (*Provider, error) {
	return &Provider{
		c: *c,
	}, nil
}

func containsOrg(orgs []string, name string) bool {
	name = strings.ToLower(name)
	for _, o := range orgs {
		if strings.ToLower(o) == name {
			return true
		}
	}
	return false
}
