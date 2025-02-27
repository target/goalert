package app

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/target/goalert/app/csp"
)

// Manually calculated (by checking Chrome console) hashes for riverui styles and scripts.
var riverStyleHashes = []string{
	"'sha256-GNF74DLkXb0fH3ILHgILFjk1ozCF3SNXQ5mQb7WLu/Y='",
	"'sha256-7Ri/I+PfhgtpcL7hT4A0VJKI6g3pK0ZvIN09RQV4ZhI='",
	"'sha256-58jqDtherY9NOM+ziRgSqQY0078tAZ+qtTBjMgbM9po='",
	"'sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU='",
	"'sha256-Nqnn8clbgv+5l0PgxcTOldg8mkMKrFn4TvPL+rYUUGg='",
	"'sha256-13vrThxdyT64GcXoTNGVoRRoL0a7EGBmOJ+lemEWyws='",
	"'sha256-QZ52fjvWgIOIOPr+gRIJZ7KjzNeTBm50Z+z9dH4N1/8='",
	"'sha256-yOU6eaJ75xfag0gVFUvld5ipLRGUy94G17B1uL683EU='",
	"'sha256-OpTmykz0m3o5HoX53cykwPhUeU4OECxHQlKXpB0QJPQ='",
	"'sha256-SSIM0kI/u45y4gqkri9aH+la6wn2R+xtcBj3Lzh7qQo='",
	"'sha256-ZH/+PJIjvP1BctwYxclIuiMu1wItb0aasjpXYXOmU0Y='",
}

var riverScriptHashes = []string{
	"'sha256-3wG3yKZPB9onV8sOQ+RPowFfmO7c5KIabwSC+UcggGo='",
}

func withSecureHeaders(enabled, https bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if !enabled {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			h := w.Header()
			if https {
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}

			nonce := uuid.NewString()
			var styleHashes, scriptHashes string
			if strings.HasPrefix(req.URL.Path, "/admin/riverui/") {
				styleHashes = strings.Join(riverStyleHashes, " ")
				scriptHashes = strings.Join(riverScriptHashes, " ")
			}

			cspVal := fmt.Sprintf("default-src 'self'; "+
				"style-src 'self' 'nonce-%s' %s;"+
				"font-src 'self' data:; "+
				"object-src 'none'; "+
				"media-src 'none'; "+
				"img-src 'self' data: https://gravatar.com/avatar/; "+
				"script-src 'self' 'nonce-%s' %s;", nonce, styleHashes, nonce, scriptHashes)

			h.Set("Content-Security-Policy", cspVal)

			h.Set("Referrer-Policy", "same-origin")
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("X-XSS-Protection", "1; mode=block")

			next.ServeHTTP(w, req.WithContext(csp.WithNonce(req.Context(), nonce)))
		})
	}
}
