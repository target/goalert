package smtpsrv

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/auth/authtoken"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"

	"github.com/emersion/go-smtp"
)

var Handler ingressHandler

type Config struct {
	Domain         string
	AllowedDomains []string
	TLSConfig      *tls.Config
}

func (cfg *Config) validDomain(d string) bool {
	if len(cfg.AllowedDomains) == 0 {
		return true
	}
	for _, v := range cfg.AllowedDomains {
		if v == d {
			return true
		}
	}
	return false
}

func NewServer(cfg *Config) *smtp.Server {
	be := new(Backend)
	s := smtp.NewServer(be)

	if cfg.Domain == "" && len(cfg.AllowedDomains) > 0 {
		cfg.Domain = cfg.AllowedDomains[0]
	}
	s.Domain = cfg.Domain

	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.AuthDisabled = true
	s.TLSConfig = cfg.TLSConfig

	return s
}

type IngressHandler interface {
	ServeSMTP(ctx context.Context, s *smtp.Server, l net.Listener)
}

type ingressHandler struct {
	alerts  *alert.Store
	intKeys IntKeyStore
	cfg     *Config
}

type IntKeyStore interface {
	Authorize(ctx context.Context, tok authtoken.Token, typ integrationkey.Type) (context.Context, error)
}

type integrationKeyStore struct {
	iks *integrationkey.Store
}

func (i *integrationKeyStore) Authorize(ctx context.Context, tok authtoken.Token, typ integrationkey.Type) (context.Context, error) {
	return i.iks.Authorize(ctx, tok, typ)
}

func (h ingressHandler) ServeSMTP(ctx context.Context, s *smtp.Server, l net.Listener) {
	err := s.Serve(l)
	if err != nil {
		log.Log(ctx, errors.New("start SMTP ingress server: "+err.Error()))
	}
}

func IngressSMTP(aDB *alert.Store, intDB *integrationkey.Store, cfg *Config) IngressHandler {
	Handler.alerts = aDB
	Handler.intKeys = &integrationKeyStore{iks: intDB}
	Handler.cfg = cfg
	return Handler
}

type Backend struct{}

func (bkd *Backend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &Session{}, nil
}

type Session struct {
	auth bool
}

func (s *Session) AuthPlain(username, password string) error {
	// log.Logf(context.Background(), "smtp auth called for user:"+username)
	s.auth = true
	return nil
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	// log.Logf(context.Background(), "Mail from:"+from)
	return nil
}

func (s *Session) Rcpt(recipient string) error {
	log.Logf(context.Background(), "Rcpt to:"+recipient)
	return nil
}

// Data is called when a new SMTP message is received
// This is the main entry point for the SMTP ingress
func (s *Session) Data(r io.Reader) error {
	ctx := context.Background()

	m, err := mail.ReadMessage(r)
	if err != nil {
		log.Log(ctx, err)
		return err
	}

	body, err := ParseSanitizeMessage(m)
	if err != nil {
		log.Log(ctx, err)
		return err
	}

	recipient := m.Header.Get("To")

	a, err := mail.ParseAddress(recipient)
	if err != nil {
		err = validation.NewFieldError("recipient", "must be valid email: "+err.Error())
	}

	recipient = a.Address

	ctx = log.WithFields(ctx, log.Fields{
		"Recipient":   recipient,
		"FromAddress": m.Header.Get("From"),
	})

	// split address
	parts := strings.SplitN(recipient, "@", 2)
	domain := strings.ToLower(parts[1])
	if !Handler.cfg.validDomain(domain) {
		err = validation.NewFieldError("domain", "invalid domain")
		log.Log(ctx, err)
		return err
	}

	// support for dedup key
	parts = strings.SplitN(parts[0], "+", 2)
	err = validate.UUID("recipient", parts[0])
	if err != nil {
		err = validation.NewFieldError("recipient", "bad mailbox name")
		log.Log(ctx, err)
		return err
	}

	tokID, err := uuid.Parse(parts[0])
	if err != nil {
		return err
	}

	tok := authtoken.Token{ID: tokID}
	var dedupStr string
	if len(parts) > 1 {
		dedupStr = parts[1]
	}

	ctx = log.WithField(ctx, "IntegrationKey", tok.ID.String())

	summary := validate.SanitizeText(m.Header.Get("Subject"), alert.MaxSummaryLength)
	details := fmt.Sprintf("From: %s\n\n%s", m.Header.Get("From"), body)
	details = validate.SanitizeText(details, alert.MaxDetailsLength)
	newAlert := &alert.Alert{
		Summary: summary,
		Details: details,
		Status:  alert.StatusTriggered,
		Source:  alert.SourceEmail,
		Dedup:   alert.NewUserDedup(dedupStr),
	}

	err = retry.DoTemporaryError(func(_ int) error {
		if newAlert.ServiceID == "" {
			ctx, err = Handler.intKeys.Authorize(ctx, tok, integrationkey.TypeEmail)
			newAlert.ServiceID = permission.ServiceID(ctx)
		}
		if err != nil {
			return err
		}
		_, _, err = Handler.alerts.CreateOrUpdate(ctx, newAlert)
		err = errors.Wrap(err, "create/update alert")
		err = errutil.MapDBError(err)
		return err
	},
		retry.Log(ctx),
		retry.Limit(12),
		retry.FibBackoff(time.Second),
	)

	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return nil
}
