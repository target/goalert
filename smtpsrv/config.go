package smtpsrv

import (
	"context"
	"crypto/tls"

	"github.com/target/goalert/alert"
)

type Config struct {
	Domain         string
	AllowedDomains []string
	TLSConfig      *tls.Config

	AuthorizeFunc   func(ctx context.Context, id string) (context.Context, error)
	CreateAlertFunc func(ctx context.Context, a *alert.Alert) error
}
