package twilio

import (
	"crypto/hmac"
	"net/http"
	"net/url"
	"regexp"

	"github.com/target/goalert/config"
	"github.com/target/goalert/util/log"

	"github.com/pkg/errors"
)

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

	var reqURL string
	if cfg.ShouldUsePublicURL() {
		reqURL = config.RequestURL(req)
	} else {
		u, err := url.ParseRequestURI(req.RequestURI)
		if err != nil {
			return err
		}
		u.Host = req.Host
		u.Scheme = req.URL.Scheme
		reqURL = u.String()
	}

	calcSig := Signature(cfg.Twilio.AuthToken, reqURL, req.PostForm)
	if !hmac.Equal([]byte(sig), calcSig) {
		return errors.New("invalid X-Twilio-Signature")
	}

	return nil
}

// WrapValidation will wrap an http.Handler to do X-Twilio-Signature checking.
func WrapValidation(h http.Handler, c Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		err := validateRequest(req)
		if err != nil {
			log.Log(ctx, err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		h.ServeHTTP(w, req)
	})
}

var (
	numRx = regexp.MustCompile(`^\+\d{1,15}$`)
	sidRx = regexp.MustCompile(`^(CA|SM)[\da-f]{32}$`)
)

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
