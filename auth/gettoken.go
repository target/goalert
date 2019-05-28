package auth

import (
	"net/http"
	"strings"
)

// GetToken will return the auth token associated with a request.
//
// Supported options (in priority order):
// - `token` (field or query)
// - Authorization: Bearer header
func GetToken(req *http.Request) string {
	tok := req.FormValue("token")
	if tok != "" {
		return tok
	}

	// compat
	tok = req.FormValue("integrationKey")
	if tok != "" {
		return tok
	}

	// compat
	tok = req.FormValue("integration_key")
	if tok != "" {
		return tok
	}

	// compat
	tok = req.FormValue("key")
	if tok != "" {
		return tok
	}

	tok = strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	if tok != "" {
		return tok
	}

	// compat
	_, tok, _ = req.BasicAuth()
	if tok != "" {
		return tok
	}

	return ""
}
