package smtpsrv

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	_ "mime"
	_ "net/http"
	_ "net/mail"
	_ "strings"
	"time"

	_ "github.com/google/uuid"
	"github.com/target/goalert/alert"
	_ "github.com/target/goalert/auth/authtoken"
	_ "github.com/target/goalert/config"
	"github.com/target/goalert/integrationkey"
	_ "github.com/target/goalert/integrationkey"
	_ "github.com/target/goalert/permission"
	_ "github.com/target/goalert/retry"
	_ "github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
	_ "github.com/target/goalert/validation"
	_ "github.com/target/goalert/validation/validate"

	"github.com/emersion/go-smtp"
)

type Config struct {
	Domain         string
	AllowedDomains []string
	ListenAddr     string
	TLSConfig      *tls.Config
}

type Backend struct{}

func (bkd *Backend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &Session{}, nil
}

type Session struct {
	auth bool
}

func (s *Session) AuthPlain(username, password string) error {
	log.Logf(context.Background(), "smtp auth called for user:", username)
	s.auth = true
	return nil
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	log.Logf(context.Background(), "Mail from:", from)
	return nil
}

func (s *Session) Rcpt(to string) error {
	log.Logf(context.Background(), "Rcpt to:", to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	if b, err := ioutil.ReadAll(r); err != nil {
		return err
	} else {
		log.Logf(context.Background(), "Data:", string(b))
	}
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
	s.Domain = cfg.Domain
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.AuthDisabled = true
	s.TLSConfig = cfg.TLSConfig

	return s
}

//
//
//
//
//
//
//
//

type ingressHandler struct {
	alerts  *alert.Store
	intKeys *integrationkey.Store
}

// func (h *ingressHandler) ServeSMTP() {
// 	ctx :=
// }

// func (h *ingressHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	cfg := config.FromContext(ctx)
// 	if !cfg.Mailgun.Enable {
// 		http.Error(w, "not enabled", http.StatusServiceUnavailable)
// 		return
// 	}

// 	ct := r.Header.Get("Content-Type")
// 	// RFC 7231, section 3.1.1.5 - empty type
// 	//   MAY be treated as application/octet-stream
// 	if ct == "" {
// 		ct = "application/octet-stream"
// 	}
// 	typ, _, err := mime.ParseMediaType(ct)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusNotAcceptable)
// 		return
// 	}

// 	switch typ {
// 	case "application/x-www-form-urlencoded":
// 		err = r.ParseForm()
// 	case "multipart/form-data", "multipart/mixed":
// 		err = r.ParseMultipartForm(32 << 20)
// 	}
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusNotAcceptable)
// 		return
// 	}

// 	// if !validSignature(ctx, r, cfg.Mailgun.APIKey) {
// 	// 	log.Log(ctx, errors.New("invalid Mailgun signature"))
// 	// 	auth.Delay(ctx)
// 	// 	http.Error(w, "Invalid Signature", http.StatusNotAcceptable)
// 	// 	return
// 	// }

// 	recipient := r.FormValue("recipient")

// 	m, err := mail.ParseAddress(recipient)
// 	if err != nil {
// 		err = validation.NewFieldError("recipient", "must be valid email: "+err.Error())
// 	}
// 	// if httpError(ctx, w, err) {
// 	// 	return
// 	// }
// 	recipient = m.Address

// 	ctx = log.WithFields(ctx, log.Fields{
// 		"Recipient":   recipient,
// 		"FromAddress": r.FormValue("from"),
// 	})

// 	// split address
// 	parts := strings.SplitN(recipient, "@", 2)
// 	domain := strings.ToLower(parts[1])
// 	if domain != cfg.Mailgun.EmailDomain {
// 		// log error? and return
// 		// httpError(ctx, w, validation.NewFieldError("domain", "invalid domain"))
// 		return
// 	}

// 	// support for dedup key
// 	parts = strings.SplitN(parts[0], "+", 2)
// 	err = validate.UUID("recipient", parts[0])
// 	if httpError(ctx, w, errors.Wrap(err, "bad mailbox name")) {
// 		return
// 	}

// 	tokID, err := uuid.Parse(parts[0])
// 	if httpError(ctx, w, err) {
// 		return
// 	}

// 	tok := authtoken.Token{ID: tokID}
// 	var dedupStr string
// 	if len(parts) > 1 {
// 		dedupStr = parts[1]
// 	}

// 	ctx = log.WithField(ctx, "IntegrationKey", tok.ID.String())

// 	summary := validate.SanitizeText(r.FormValue("subject"), alert.MaxSummaryLength)
// 	details := fmt.Sprintf("From: %s\n\n%s", r.FormValue("from"), r.FormValue("body-plain"))
// 	details = validate.SanitizeText(details, alert.MaxDetailsLength)
// 	newAlert := &alert.Alert{
// 		Summary: summary,
// 		Details: details,
// 		Status:  alert.StatusTriggered,
// 		Source:  alert.SourceEmail,
// 		Dedup:   alert.NewUserDedup(dedupStr),
// 	}

// 	err = retry.DoTemporaryError(func(_ int) error {
// 		if newAlert.ServiceID == "" {
// 			ctx, err = h.intKeys.Authorize(ctx, tok, integrationkey.TypeEmail)
// 			newAlert.ServiceID = permission.ServiceID(ctx)
// 		}
// 		if err != nil {
// 			return err
// 		}
// 		_, err = h.alerts.CreateOrUpdate(ctx, newAlert)
// 		err = errors.Wrap(err, "create/update alert")
// 		err = errutil.MapDBError(err)
// 		return err
// 	},
// 		retry.Log(ctx),
// 		retry.Limit(12),
// 		retry.FibBackoff(time.Second),
// 	)

// 	httpError(ctx, w, err)
// }
