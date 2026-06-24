package app

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/target/goalert/app/csp"
)

// Manually calculated (by checking dev console) hashes for riverui styles and scripts.
var riverStyleHashes = []string{
	"'sha256-dd4J3UnQShsOmqcYi4vN5BT3mGZB/0fOwBA72rsguKc='",
	"'sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU='",
	"'sha256-Nqnn8clbgv+5l0PgxcTOldg8mkMKrFn4TvPL+rYUUGg='",
	"'sha256-13vrThxdyT64GcXoTNGVoRRoL0a7EGBmOJ+lemEWyws='",
	"'sha256-QZ52fjvWgIOIOPr+gRIJZ7KjzNeTBm50Z+z9dH4N1/8='",
	"'sha256-yOU6eaJ75xfag0gVFUvld5ipLRGUy94G17B1uL683EU='",
	"'sha256-OpTmykz0m3o5HoX53cykwPhUeU4OECxHQlKXpB0QJPQ='",
	"'sha256-SSIM0kI/u45y4gqkri9aH+la6wn2R+xtcBj3Lzh7qQo='",
	"'sha256-ZH/+PJIjvP1BctwYxclIuiMu1wItb0aasjpXYXOmU0Y='",
	"'sha256-58jqDtherY9NOM+ziRgSqQY0078tAZ+qtTBjMgbM9po='",
	"'sha256-7Ri/I+PfhgtpcL7hT4A0VJKI6g3pK0ZvIN09RQV4ZhI='",
	"'sha256-GNF74DLkXb0fH3ILHgILFjk1ozCF3SNXQ5mQb7WLu/Y='",
	"'sha256-skqujXORqzxt1aE0NNXxujEanPTX6raoqSscTV/Ww/Y='",
	"'sha256-x8oKdtSwwf2MHmRCE1ArEPR/R4NRjiMqSu6isbLZIUo='",
	"'sha256-MDf+R0QbM9MuKMsR2e99weO3pEauOCVCpaP4bsB8KRg='",
}

var riverScriptHashes = []string{
	"'sha256-9IKZGijA20+zzz3VIneuNo2k1OVkHiiOk2VKTKZjqLc='",
	"'sha256-FhazKW7/4VRAybIf+mFprqYHfRXCMp1Rqh1PhpxSwtk='",
	"'sha256-/c0mqg4UDO/IaoMY9uypUqf4nzFpiLMms1Gcdr2XqcU='",
	"'sha256-4o5fFgJhRFoLYxAPc5xSpNr7R53Z3QEJ+2XnHXOVrJ8='",
	"'sha256-xUpbdveMn6brc/ivPFp80kPtDiPVhWwS7FJ2B4HkME0='",
	"'sha256-WHOj9nkTdv7Fqj4KfdVoW0fBeUZRTjCoKeSgjjf33uc='",
	"'sha256-kwnxJYYglj1d+/ZNxVOqRpRK80ZYeddMAIosyubwDXI='",
	"'sha256-iOOYu2PDgIl6ATjPEoSJrzHdRadFMG4Nyc7hNqwsc3U='",
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
			var cspVal string
			if strings.HasPrefix(req.URL.Path, "/admin/riverui/") {
				// Until RiverUI fully supports CSP, we need to allow its inline styles and scripts.
				// This is done by including the hashes of the inline styles/scripts used in RiverUI.
				// These hashes are manually calculated by checking the dev console.
				styleHashes := strings.Join(riverStyleHashes, " ")
				scriptHashes := strings.Join(riverScriptHashes, " ")
				cspVal = fmt.Sprintf("default-src 'self'; "+
					"style-src 'self' 'nonce-%s' %s;"+
					"font-src 'self' data:; "+
					"object-src 'none'; "+
					"media-src 'none'; "+
					"img-src 'self' data: https://gravatar.com/avatar/; "+
					"script-src 'self' 'unsafe-eval' 'nonce-%s' %s;", nonce, styleHashes, nonce, scriptHashes)
			} else {
				cspVal = fmt.Sprintf("default-src 'self'; "+
					"style-src 'self' 'nonce-%s';"+
					"font-src 'self' data:; "+
					"object-src 'none'; "+
					"media-src 'none'; "+
					"img-src 'self' data: https://gravatar.com/avatar/; "+
					"script-src 'self' 'nonce-%s';", nonce, nonce)
			}

			h.Set("Content-Security-Policy", cspVal)

			h.Set("Referrer-Policy", "same-origin")
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("X-XSS-Protection", "1; mode=block")

			next.ServeHTTP(w, req.WithContext(csp.WithNonce(req.Context(), nonce)))
		})
	}
}
