package auth

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/target/goalert/config"
)

type disableSameSiteKey struct{}

// WithDisableSameSite will return a new context with the disableSameSiteKey set.
func WithDisableSameSite(ctx context.Context) context.Context {
	return context.WithValue(ctx, disableSameSiteKey{}, true)
}

func isSameSiteDisabled(ctx context.Context) bool {
	return ctx.Value(disableSameSiteKey{}) != nil
}

// SetCookie will set a cookie value for all API prefixes, respecting the current config parameters.
func SetCookie(w http.ResponseWriter, req *http.Request, name, value string, isSession bool) {
	SetCookieAge(w, req, name, value, 0, isSession)
}

// SetCookieAge behaves like SetCookie but also sets the MaxAge.
func SetCookieAge(w http.ResponseWriter, req *http.Request, name, value string, age time.Duration, isSession bool) {
	cfg := config.FromContext(req.Context())
	u, err := url.Parse(cfg.PublicURL())
	if err != nil {
		panic(err)
	}

	cookiePath := "/"
	secure := req.URL.Scheme == "https"
	if cfg.ShouldUsePublicURL() {
		cookiePath = strings.TrimSuffix(u.Path, "/") + "/"
		secure = u.Scheme == "https"
	}

	// Use Lax mode for non-session cookies, this allows the cookie to be sent when
	// navigating to the login page from a different domain (e.g., OAuth redirect).
	sameSite := http.SameSiteLaxMode
	if isSession {
		if isSameSiteDisabled(req.Context()) {
			sameSite = http.SameSiteNoneMode
		} else {
			sameSite = http.SameSiteStrictMode
		}
	}

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Secure:   secure,
		Name:     name,

		Path:   cookiePath,
		Value:  value,
		MaxAge: int(age.Seconds()),

		SameSite: sameSite,
	})
}

// ClearCookie will clear and expire the cookie with the given name, for all API prefixes.
func ClearCookie(w http.ResponseWriter, req *http.Request, name string, isSession bool) {
	SetCookieAge(w, req, name, "", -time.Second, isSession)
}
