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
	"github.com/target/goalert/alert"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/validation/validate"
)

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

func (s *Session) AuthPlain(username, password string) error { return smtp.ErrAuthUnsupported }

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	addr, err := mail.ParseAddress(from)
	if err != nil {
		return err
	}

	s.from = addr.String()
	return nil
}

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

	m, err := mail.ReadMessage(r)
	if err != nil {
		return err
	}

	body, err := ParseSanitizeMessage(m)
	if err != nil {
		return err
	}

	summary := validate.SanitizeText(m.Header.Get("Subject"), alert.MaxSummaryLength)
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

func (s *Session) Reset() {
	s.dedup = ""
	s.from = ""
	s.authCtx = nil
}

func (s *Session) Logout() error { return nil }
