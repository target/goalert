package graphqlapp

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuth state storage (in production, this should be stored in Redis or DB)
var oauthStates = make(map[string]*oauthState)

type oauthState struct {
	ClientID     string
	ClientSecret string
	CreatedAt    time.Time
}

// Mutation resolvers for IMAP OAuth

func (m *Mutation) GenerateIMAPOAuthURL(ctx context.Context, input graphql2.GenerateIMAPOAuthURLInput) (*graphql2.IMAPOAuthURL, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	// Generate random state token
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return nil, fmt.Errorf("failed to generate state token: %w", err)
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Store state with OAuth credentials
	oauthStates[state] = &oauthState{
		ClientID:     input.ClientID,
		ClientSecret: input.ClientSecret,
		CreatedAt:    time.Now(),
	}

	// Clean up old states (older than 10 minutes)
	for k, v := range oauthStates {
		if time.Since(v.CreatedAt) > 10*time.Minute {
			delete(oauthStates, k)
		}
	}

	// Create OAuth config
	config := &oauth2.Config{
		ClientID:     input.ClientID,
		ClientSecret: input.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  input.RedirectURL,
		Scopes:       []string{"https://mail.google.com/"}, // Gmail IMAP scope
	}

	// Generate auth URL
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	return &graphql2.IMAPOAuthURL{
		AuthURL: authURL,
		State:   state,
	}, nil
}

func (m *Mutation) ExchangeIMAPOAuthCode(ctx context.Context, input graphql2.ExchangeIMAPOAuthCodeInput) (*graphql2.IMAPOAuthToken, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	// Retrieve stored OAuth state
	storedState, ok := oauthStates[input.State]
	if !ok {
		return nil, fmt.Errorf("invalid or expired state token")
	}

	// Clean up state
	delete(oauthStates, input.State)

	// Check if state is not too old
	if time.Since(storedState.CreatedAt) > 10*time.Minute {
		return nil, fmt.Errorf("state token expired")
	}

	// Create OAuth config
	config := &oauth2.Config{
		ClientID:     storedState.ClientID,
		ClientSecret: storedState.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  input.RedirectURL,
		Scopes:       []string{"https://mail.google.com/"},
	}

	// Exchange code for token
	token, err := config.Exchange(context.Background(), input.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	if token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token received - you may need to revoke access and try again")
	}

	return &graphql2.IMAPOAuthToken{
		RefreshToken: token.RefreshToken,
		AccessToken:  token.AccessToken,
		ExpiresAt:    token.Expiry,
	}, nil
}
