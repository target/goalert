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

var Handler IngressHandler

type Config struct {
	Domain         string
	AllowedDomains []string
	ListenAddr     string
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

type Backend struct{}

func (bkd *Backend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &Session{}, nil
}

type Session struct {
	auth bool
}

func (s *Session) AuthPlain(username, password string) error {
	log.Logf(context.Background(), "smtp auth called for user:"+username)
	s.auth = true
	return nil
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	log.Logf(context.Background(), "Mail from:"+from)
	return nil
}

func (s *Session) Rcpt(recipient string) error {
	log.Logf(context.Background(), "Rcpt to:"+recipient)
	m, err := mail.ParseAddress(recipient)
	if err != nil {
		err = validation.NewFieldError("recipient", "must be valid email: "+err.Error())
		log.Log(context.Background(), err)
	}

	recipient = m.Address
	log.Logf(context.Background(), "recipient = "+recipient)
	return nil
}

// Data is called when a new SMTP message is received
func (s *Session) Data(r io.Reader) error {
	ctx := context.Background()

	m, err := mail.ReadMessage(r)
	if err != nil {
		log.Log(ctx, err)
		return nil
	}

	logMsg := `Date: %s
From: %s
To: %s
Subject: %s

%s
`

	header := m.Header
	body, err := io.ReadAll(m.Body)
	if err != nil {
		log.Log(ctx, err)
		return nil
	}
	log.Logf(ctx, logMsg, header.Get("Date"), header.Get("From"), header.Get("To"), header.Get("Subject"), string(body))

	recipient := header.Get("To")

	a, err := mail.ParseAddress(recipient)
	if err != nil {
		err = validation.NewFieldError("recipient", "must be valid email: "+err.Error())
	}

	recipient = a.Address

	ctx = log.WithFields(ctx, log.Fields{
		"Recipient":   recipient,
		"FromAddress": header.Get("From"),
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

	summary := validate.SanitizeText(header.Get("Subject"), alert.MaxSummaryLength)
	details := fmt.Sprintf("From: %s\n\n%s", header.Get("From"), body)
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
		_, err = Handler.alerts.CreateOrUpdate(ctx, newAlert)
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

func NewServer(cfg *Config) *smtp.Server {
	be := new(Backend)
	s := smtp.NewServer(be)

	s.Addr = cfg.ListenAddr
	fmt.Printf("creating new SMTP server on addr %s\n", s.Addr)

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

type IngressHandler struct {
	alerts  *alert.Store
	intKeys *integrationkey.Store
	cfg     *Config
}

func (H *IngressHandler) ServeSMTP(ctx context.Context, s *smtp.Server, l net.Listener) {
	err := s.Serve(l)
	if err != nil {
		log.Log(ctx, errors.New("start SMTP receiver server"))
	}

}

func IngressSMTP(aDB *alert.Store, intDB *integrationkey.Store, cfg *Config) IngressHandler {
	Handler.alerts = aDB
	Handler.intKeys = intDB
	Handler.cfg = cfg
	return Handler
}
