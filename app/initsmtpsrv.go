package app

import (
	"context"
	"crypto/tls"
	"net"
	_ "net/url"
	"strings"

	"github.com/target/goalert/smtpsrv"
)

func (app *App) initSMTPServer(ctx context.Context) error {
	cfg := smtpsrv.Config{}

	if app.cfg.SMTPListenAddr == "" && app.cfg.SMTPListenAddrTLS == "" {
		return nil
	}

	if app.cfg.SMTPAllowedDomains != "" {
		cfg.AllowedDomains = strings.Split(app.cfg.SMTPAllowedDomains, ",")
	}

	cfg.TLSConfig = app.cfg.TLSConfigSMTP // nil if unset

	app.smtpsrv = smtpsrv.NewServer(&cfg)
	h := smtpsrv.IngressSMTP(app.AlertStore, app.IntegrationKeyStore, &cfg)

	if app.cfg.SMTPListenAddr != "" {
		l, err := net.Listen("tcp", app.cfg.SMTPListenAddr)
		if err != nil {
			return err
		}
		go func() {
			h.ServeSMTP(ctx, app.smtpsrv, l)
		}()
	}

	if app.cfg.SMTPListenAddrTLS != "" {
		l, err := tls.Listen("tcp", app.cfg.SMTPListenAddrTLS, cfg.TLSConfig)
		if err != nil {
			return err
		}
		go func() {
			h.ServeSMTP(ctx, app.smtpsrv, l)
		}()
	}
	return nil
}
