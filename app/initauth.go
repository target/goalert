package app

import (
	"context"

	"github.com/pkg/errors"
	"github.com/target/goalert/auth"
	"github.com/target/goalert/auth/basic"
	"github.com/target/goalert/auth/github"
	"github.com/target/goalert/auth/oidc"
)

func (app *App) initAuth(ctx context.Context) error {

	var err error
	app.AuthHandler, err = auth.NewHandler(ctx, app.db, auth.HandlerConfig{
		UserStore:      app.UserStore,
		SessionKeyring: app.SessionKeyring,
		IntKeyStore:    app.IntegrationKeyStore,
		CalSubStore:    app.CalSubStore,
		APIKeyring:     app.APIKeyring,
	})
	if err != nil {
		return errors.Wrap(err, "init auth handler")
	}

	cfg := oidc.Config{
		Keyring:    app.OAuthKeyring,
		NonceStore: app.NonceStore,
	}
	oidcProvider, err := oidc.NewProvider(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "init OIDC auth provider")
	}
	if err := app.AuthHandler.AddIdentityProvider("oidc", oidcProvider); err != nil {
		return err
	}

	githubConfig := &github.Config{
		Keyring:    app.OAuthKeyring,
		NonceStore: app.NonceStore,
	}

	githubProvider, err := github.NewProvider(ctx, githubConfig)
	if err != nil {
		return errors.Wrap(err, "init GitHub auth provider")
	}
	if err := app.AuthHandler.AddIdentityProvider("github", githubProvider); err != nil {
		return err
	}

	basicProvider, err := basic.NewProvider(ctx, app.AuthBasicStore)
	if err != nil {
		return errors.Wrap(err, "init basic auth provider")
	}
	return app.AuthHandler.AddIdentityProvider("basic", basicProvider)
}
