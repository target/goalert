package twilio

import (
	"io"
	"net/http"

	"github.com/felixge/httpsnoop"
)

// WrapHeaderHack wraps an http.Handler so that a 204 is returned if the body is empty.
//
// A Go 1.10 change removed the implicit header for responses with no content. Unfortunately
// Twilio logs empty responses (with no `Content-Type`) as 502s.
func WrapHeaderHack(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var wrote bool
		ww := httpsnoop.Wrap(w, httpsnoop.Hooks{
			Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
				wrote = true
				return next
			},
			WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
				wrote = true
				return next
			},
			ReadFrom: func(next httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
				wrote = true
				return func(src io.Reader) (int64, error) {
					n, err := next(src)
					if n > 0 {
						wrote = true
					}
					return n, err
				}
			},
		})

		h.ServeHTTP(ww, req)

		if !wrote {
			w.WriteHeader(204)
		}
	})
}
