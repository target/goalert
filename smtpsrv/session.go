package smtpsrv

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"strings"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/mnako/letters"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/validation/validate"
)

// Session implements an SMTP session that creates alerts.
type Session struct {
	cfg Config

	from    string
	dedup   string
	authCtx context.Context
}

func (s *Session) isValidDomain(d string) bool {
	if strings.EqualFold(d, s.cfg.Domain) {
		return true
	}

	for _, v := range s.cfg.AllowedDomains {
		if strings.EqualFold(v, d) {
			return true
		}
	}

	return false
}

// AuthPlain is called when a client attempts to authenticate using the PLAIN
// auth mechanism. It always returns an error, indicating that PLAIN auth is
// not supported.
func (s *Session) AuthPlain(username, password string) error { return smtp.ErrAuthUnsupported }

// Mail is called when a new SMTP message is received (MAIL FROM). It checks
// that the sender is valid and stores it in the session.
func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	addr, err := mail.ParseAddress(from)
	if err != nil {
		return err
	}

	s.from = addr.String()
	return nil
}

// Rcpt is called when a new SMTP message is received (RCPT TO). It checks
// that the recipient is valid and stores it in the session.
//
// It also checks that the recipient is authorized to create alerts.
func (s *Session) Rcpt(recipient string) error {
	addr, err := mail.ParseAddress(recipient)
	if err != nil {
		return err
	}
	id, domain, ok := strings.Cut(addr.Address, "@")
	if !ok {
		return errors.New("invalid recipient")
	}
	if !s.isValidDomain(domain) {
		return errors.New("invalid domain")
	}
	id, s.dedup, _ = strings.Cut(id, "+")
	err = validate.UUID("recipient", id)
	if err != nil {
		return err
	}

	ctx, err := s.cfg.AuthorizeFunc(context.Background(), id)
	if err != nil {
		return err
	}

	s.authCtx = ctx
	return nil
}

// Data is called when a new SMTP message is received.
func (s *Session) Data(r io.Reader) error {
	if s.authCtx == nil {
		return errors.New("no recipient")
	}

	email, err := letters.ParseEmail(r)
	if err != nil {
		return err
	}
	body := email.Text

	summary := validate.SanitizeText(email.Headers.Subject, alert.MaxSummaryLength)
	details := fmt.Sprintf("From: %s\n\n%s", s.from, body)
	details = validate.SanitizeText(details, alert.MaxDetailsLength)
	var dedup *alert.DedupID
	if s.dedup != "" {
		dedup = alert.NewUserDedup(s.dedup)
	}
	newAlert := &alert.Alert{
		Summary:   summary,
		Details:   details,
		ServiceID: permission.ServiceID(s.authCtx),
		Status:    alert.StatusTriggered,
		Source:    alert.SourceEmail,
		Dedup:     dedup,
	}

	return retry.DoTemporaryError(func(_ int) error {
		return s.cfg.CreateAlertFunc(s.authCtx, newAlert)
	},
		retry.Log(s.authCtx),
		retry.Limit(12),
		retry.FibBackoff(time.Second),
	)
}

// Reset resets the session state.
func (s *Session) Reset() {
	s.dedup = ""
	s.from = ""
	s.authCtx = nil
}

// Logout is called when the client requests to log out.
func (s *Session) Logout() error { return nil }
