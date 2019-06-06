package auth

import (
	"net/http"
	"time"
)

// SetCookie will set a cookie value for all API prefixes, respecting the current config parameters.
func SetCookie(w http.ResponseWriter, req *http.Request, name, value string) {
	SetCookieAge(w, req, name, value, 0)
}

// SetCookieAge behaves like SetCookie but also sets the MaxAge.
func SetCookieAge(w http.ResponseWriter, req *http.Request, name, value string, age time.Duration) {
	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Secure:   req.URL.Scheme == "https",
		Name:     name,

		// Until we can finish removing /v1 from all UI calls
		// we need cookies available on both /api and /v1.
		//
		// Unfortunately we can't just set both paths without breaking integration tests...
		// We'll keep this as `/` until Cypress fixes it's cookie handling, or we
		// finish removing the `/v1` UI code. Whichever is sooner.
		Path:   "/",
		Value:  value,
		MaxAge: int(age.Seconds()),
	})
}

// ClearCookie will clear and expire the cookie with the given name, for all API prefixes.
func ClearCookie(w http.ResponseWriter, req *http.Request, name string) {
	SetCookieAge(w, req, name, "", -time.Second)
}
