package twilio

import (
	"context"
	"crypto/hmac"
	"github.com/target/goalert/config"
	"github.com/target/goalert/util/log"
	"net/http"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

type contextKey string

const twilioAlreadyValidated = contextKey("already-validated")

func validateRequest(req *http.Request) error {
	if req.Method == "POST" {
		req.ParseForm()
	}
	ctx := req.Context()
	cfg := config.FromContext(ctx)

	sig := req.Header.Get("X-Twilio-Signature")
	if sig == "" {
		return errors.New("missing X-Twilio-Signature")
	}

	calcSig := Signature(cfg.Twilio.AuthToken, cfg.CallbackURL(req.URL.String()), req.PostForm)
	if !hmac.Equal([]byte(sig), calcSig) {
		return errors.New("invalid X-Twilio-Signature")
	}

	return nil
}

// WrapValidation will wrap an http.Handler to do X-Twilio-Signature checking.
func WrapValidation(h http.Handler, c Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		if val, ok := ctx.Value(twilioAlreadyValidated).(bool); ok && val {
			// only validate once
			h.ServeHTTP(w, req)
			return
		}

		err := validateRequest(req)
		if err != nil {
			log.Log(ctx, err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		h.ServeHTTP(w, req.WithContext(context.WithValue(ctx, twilioAlreadyValidated, true)))
	})
}

var numRx = regexp.MustCompile(`^\+\d{1,15}$`)
var sidRx = regexp.MustCompile(`^(CA|SM)[\da-f]{32}$`)

func validPhone(n string) string {
	if !numRx.MatchString(n) {
		return ""
	}

	return n
}
func validSID(n string) string {
	if len(n) != 34 {
		return ""
	}
	if !sidRx.MatchString(n) {
		return ""
	}

	return n
}

// Supported Country Codes
// +1 = USA, +91 = India, +44 = United Kingdom
func supportedCountryCode(n string) bool {
	if strings.HasPrefix(n, "+1") || strings.HasPrefix(n, "+91") || strings.HasPrefix(n, "+44") {
		return true
	}
	return false
}
