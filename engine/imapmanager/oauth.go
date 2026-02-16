package imapmanager

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// getOAuthAccessTokenForService retrieves a fresh OAuth access token for a specific service.
func (db *DB) getOAuthAccessTokenForService(ctx context.Context, svc IMAPServiceConfig) (string, error) {
	// Check if OAuth is configured for this service
	if !svc.OAuthClientID.Valid || !svc.OAuthClientSecret.Valid || !svc.OAuthRefreshToken.Valid {
		return "", fmt.Errorf("OAuth not configured for service: missing ClientID, ClientSecret, or RefreshToken")
	}

	// Create OAuth2 config for Gmail
	oauthConfig := &oauth2.Config{
		ClientID:     svc.OAuthClientID.String,
		ClientSecret: svc.OAuthClientSecret.String,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://mail.google.com/"}, // Gmail IMAP scope
	}

	// Create token from refresh token
	token := &oauth2.Token{
		RefreshToken: svc.OAuthRefreshToken.String,
	}

	// Use the token source to get a fresh access token
	tokenSource := oauthConfig.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to refresh OAuth token: %w", err)
	}

	return newToken.AccessToken, nil
}
