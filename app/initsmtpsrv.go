package app

import (
	"context"
	_ "net/url"
	"strings"

	"github.com/target/goalert/smtpsrv"
)

func (app *App) initSMTPServer(ctx context.Context) error {
	cfg := smtpsrv.Config{}

	if app.cfg.SMTPListenAddr == "" && app.cfg.SMTPListenAddrTLS == "" {
		return nil
	}

	cfg.AllowedDomains = strings.Split(app.cfg.SMTPAllowedDomains, ",")
	cfg.Domain = ""
	cfg.TLSConfig = app.cfg.TLSConfigSMTP

	if app.cfg.SMTPListenAddrTLS != "" {
		cfg.ListenAddr = app.cfg.SMTPListenAddrTLS
		s := smtpsrv.NewServer(&cfg)
		if err := s.ListenAndServeTLS(); err != nil {
			return err
		}
	} else {
		cfg.ListenAddr = app.cfg.SMTPListenAddr
		s := smtpsrv.NewServer(&cfg)
		if err := s.ListenAndServe(); err != nil {
			return err
		}
	}
	return nil
}
