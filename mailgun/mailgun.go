package mailgun

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/auth"
	"github.com/target/goalert/auth/authtoken"
	"github.com/target/goalert/config"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// httpError is used to respond in a standard way to Mailgun when err != nil. If
// err is nil, false is returned, true otherwise.
// If Mailgun receives a 200 (Success) code it will determine the webhook POST is successful and not retry.
// If Mailgun receives a 406 (Not Acceptable) code, Mailgun will determine the POST is rejected and not retry.
//
// For any other code, Mailgun will retry POSTing according to the following schedule (other than the delivery notification):
// 10 minutes, 10 minutes, 15 minutes, 30 minutes, 1 hour, 2 hour and 4 hours.
func httpError(ctx context.Context, w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	var clientErr interface {
		ClientError() bool
	}

	if errors.As(err, &clientErr) && clientErr.ClientError() {
		log.Debug(ctx, err)
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return true
	}

	if errutil.IsLimitError(err) {
		// don't retry if a limit has been exceeded
		log.Debug(ctx, err)
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return true
	}

	log.Log(ctx, err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	return true
}

// validSignature is used to validate the request from Mailgun.
// If request is validated true is returned, false otherwise.
// https://documentation.mailgun.com/en/latest/user_manual.html#securing-webhooks
func validSignature(ctx context.Context, req *http.Request, apikey string) bool {
	h := hmac.New(sha256.New, []byte(apikey))
	_, _ = io.WriteString(h, req.FormValue("timestamp"))
	_, _ = io.WriteString(h, req.FormValue("token"))

	calculatedSignature := h.Sum(nil)
	signature, err := hex.DecodeString(req.FormValue("signature"))
	if err != nil {
		return false
	}

	return hmac.Equal(signature, calculatedSignature)
}

type ingressHandler struct {
	alerts  *alert.Store
	intKeys *integrationkey.Store
}

func (h *ingressHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cfg := config.FromContext(ctx)
	if !cfg.Mailgun.Enable {
		http.Error(w, "not enabled", http.StatusServiceUnavailable)
		return
	}

	ct := r.Header.Get("Content-Type")
	// RFC 7231, section 3.1.1.5 - empty type
	//   MAY be treated as application/octet-stream
	if ct == "" {
		ct = "application/octet-stream"
	}
	typ, _, err := mime.ParseMediaType(ct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	switch typ {
	case "application/x-www-form-urlencoded":
		err = r.ParseForm()
	case "multipart/form-data", "multipart/mixed":
		err = r.ParseMultipartForm(32 << 20)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	if !validSignature(ctx, r, cfg.Mailgun.APIKey) {
		log.Log(ctx, errors.New("invalid Mailgun signature"))
		auth.Delay(ctx)
		http.Error(w, "Invalid Signature", http.StatusNotAcceptable)
		return
	}

	recipient := r.FormValue("recipient")

	m, err := mail.ParseAddress(recipient)
	if err != nil {
		err = validation.NewFieldError("recipient", "must be valid email: "+err.Error())
	}
	if httpError(ctx, w, err) {
		return
	}
	recipient = m.Address

	ctx = log.WithFields(ctx, log.Fields{
		"Recipient":   recipient,
		"FromAddress": r.FormValue("from"),
	})

	// split address
	parts := strings.SplitN(recipient, "@", 2)
	domain := strings.ToLower(parts[1])
	if domain != cfg.Mailgun.EmailDomain {
		httpError(ctx, w, validation.NewFieldError("domain", "invalid domain"))
		return
	}

	// support for dedup key
	parts = strings.SplitN(parts[0], "+", 2)
	err = validate.UUID("recipient", parts[0])
	if httpError(ctx, w, errors.Wrap(err, "bad mailbox name")) {
		return
	}

	tokID, err := uuid.Parse(parts[0])
	if httpError(ctx, w, err) {
		return
	}

	tok := authtoken.Token{ID: tokID}
	var dedupStr string
	if len(parts) > 1 {
		dedupStr = parts[1]
	}

	ctx = log.WithField(ctx, "IntegrationKey", tok.ID.String())

	summary := validate.SanitizeText(r.FormValue("subject"), alert.MaxSummaryLength)
	details := fmt.Sprintf("From: %s\n\n%s", r.FormValue("from"), r.FormValue("body-plain"))
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
			ctx, err = h.intKeys.Authorize(ctx, tok, integrationkey.TypeEmail)
			newAlert.ServiceID = permission.ServiceID(ctx)
		}
		if err != nil {
			return err
		}
		_, _, err = h.alerts.CreateOrUpdate(ctx, newAlert, nil)
		err = errors.Wrap(err, "create/update alert")
		err = errutil.MapDBError(err)
		return err
	},
		retry.Log(ctx),
		retry.Limit(12),
		retry.FibBackoff(time.Second),
	)

	httpError(ctx, w, err)
}

// IngressWebhooks is used to accept webhooks from Mailgun to support email as an alert creation mechanism.
// Will read POST form parameters, validate, sanitize and use to create a new alert.
// https://documentation.mailgun.com/en/latest/user_manual.html#parsed-messages-parameters
func IngressWebhooks(aDB *alert.Store, intDB *integrationkey.Store) http.HandlerFunc {
	return (&ingressHandler{
		alerts:  aDB,
		intKeys: intDB,
	}).ServeHTTP
}
