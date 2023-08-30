package app

import (
	"context"
	"crypto/tls"
	"net"
	_ "net/url"
	"strings"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/auth/authtoken"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/smtpsrv"
)

func (app *App) initSMTPServer(ctx context.Context) error {
	if app.cfg.SMTPListenAddr == "" && app.cfg.SMTPListenAddrTLS == "" {
		return nil
	}

	cfg := smtpsrv.Config{
		Domain:            app.cfg.EmailIntegrationDomain,
		AllowedDomains:    parseAllowedDomains(app.cfg.SMTPAdditionalDomains, app.cfg.EmailIntegrationDomain),
		TLSConfig:         app.cfg.TLSConfigSMTP,
		MaxRecipients:     app.cfg.SMTPMaxRecipients,
		BackgroundContext: app.LogBackgroundContext,
		AuthorizeFunc: func(ctx context.Context, id string) (context.Context, error) {
			tok, _, err := authtoken.Parse(id, nil)
			if err != nil {
				return nil, err
			}

			ctx, err = app.IntegrationKeyStore.Authorize(ctx, *tok, integrationkey.TypeEmail)
			if err != nil {
				return nil, err
			}

			return ctx, nil
		},
		CreateAlertFunc: func(ctx context.Context, a *alert.Alert) error {
			_, _, err := app.AlertStore.CreateOrUpdate(ctx, a)
			return err
		},
	}

	app.smtpsrv = smtpsrv.NewServer(cfg)
	var err error
	if app.cfg.SMTPListenAddr != "" {
		app.smtpsrvL, err = net.Listen("tcp", app.cfg.SMTPListenAddr)
		if err != nil {
			return err
		}
	}

	if app.cfg.SMTPListenAddrTLS != "" {
		l, err := tls.Listen("tcp", app.cfg.SMTPListenAddrTLS, cfg.TLSConfig)
		if err != nil {
			return err
		}
		app.smtpsrvL = newMultiListener(app.cfg.Logger, app.smtpsrvL, l)
	}

	return nil
}

func parseAllowedDomains(additionalDomains string, primaryDomain string) []string {
	if !strings.Contains(additionalDomains, primaryDomain) {
		additionalDomains = strings.Join([]string{additionalDomains, primaryDomain}, ",")
	}
	return strings.Split(additionalDomains, ",")
}
