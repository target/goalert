package validate

import (
	"fmt"
	"strings"

	"github.com/target/goalert/validation"
)

// OAuthScope is a valid OAuth 2.0 Scope parameter, optionally specifying a list of required scopes.
//
// Requirements based on section 3.3 of RFC6749
// https://tools.ietf.org/html/rfc6749#section-3.3
func OAuthScope(fname, val string, required ...string) error {
	if val == "" {
		return validation.NewFieldError(fname, "must not be empty")
	}
	for _, r := range val {
		if r == ' ' {
			continue
		}
		if r == 0x21 {
			continue
		}
		if r >= 0x23 && r <= 0x5b {
			continue
		}
		if r >= 0x5d && r <= 0x7e {
			continue
		}

		return validation.NewFieldError(fname, "invalid character")
	}
	scopes := strings.Split(val, " ")
	m := make(map[string]bool, len(scopes))
	for _, scope := range scopes {
		if len(scope) == 0 {
			return validation.NewFieldError(fname, "must not contain empty scopes")
		}
		if m[scope] {
			return validation.NewFieldError(fname, fmt.Sprintf("duplicate scope '%s'", scope))
		}
		m[scope] = true
	}

	for _, req := range required {
		if !m[req] {
			return validation.NewFieldError(fname, fmt.Sprintf("missing required scope '%s'", req))
		}
	}

	return nil
}
