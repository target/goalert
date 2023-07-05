package auth

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/target/goalert/config"
)

// SetCookie will set a cookie value for all API prefixes, respecting the current config parameters.
func SetCookie(w http.ResponseWriter, req *http.Request, name, value string) {
	SetCookieAge(w, req, name, value, 0)
}

// SetCookieAge behaves like SetCookie but also sets the MaxAge.
func SetCookieAge(w http.ResponseWriter, req *http.Request, name, value string, age time.Duration) {
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

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Secure:   secure,
		Name:     name,

		Path:   cookiePath,
		Value:  value,
		MaxAge: int(age.Seconds()),

		SameSite: http.SameSiteStrictMode,
	})
}

// ClearCookie will clear and expire the cookie with the given name, for all API prefixes.
func ClearCookie(w http.ResponseWriter, req *http.Request, name string) {
	SetCookieAge(w, req, name, "", -time.Second)
}
