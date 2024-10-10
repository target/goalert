package app

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/target/goalert/app/csp"
)

func withSecureHeaders(disable, https bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if disable {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			h := w.Header()
			if https {
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}

			nonce := uuid.NewString()

			cspVal := "default-src 'self'; " +
				"style-src 'self' 'nonce-" + nonce + "'; " +
				"font-src 'self' data:; " +
				"object-src 'none'; " +
				"media-src 'none'; " +
				"img-src 'self' data: https://gravatar.com/avatar/; " +
				"script-src 'self' 'nonce-" + nonce + "';"

			h.Set("Content-Security-Policy", cspVal)

			h.Set("Referrer-Policy", "same-origin")
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("X-XSS-Protection", "1; mode=block")

			next.ServeHTTP(w, req.WithContext(csp.WithNonce(req.Context(), nonce)))
		})
	}
}
