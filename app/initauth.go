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
	app.authHandler, err = auth.NewHandler(ctx, app.db, auth.HandlerConfig{
		UserStore:      app.UserStore,
		SessionKeyring: app.SessionKeyring,
		IntKeyStore:    app.IntegrationKeyStore,
		CalSubStore:    app.CalSubStore,
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
	app.authHandler.AddIdentityProvider("oidc", oidcProvider)

	githubConfig := &github.Config{
		Keyring:    app.OAuthKeyring,
		NonceStore: app.NonceStore,
	}

	githubProvider, err := github.NewProvider(ctx, githubConfig)
	if err != nil {
		return errors.Wrap(err, "init GitHub auth provider")
	}
	app.authHandler.AddIdentityProvider("github", githubProvider)

	basicProvider, err := basic.NewProvider(ctx, app.db)
	if err != nil {
		return errors.Wrap(err, "init basic auth provider")
	}
	app.authHandler.AddIdentityProvider("basic", basicProvider)

	return err
}
