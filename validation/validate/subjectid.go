package validate

import (
	"github.com/target/goalert/validation"
	"strings"
)

// SubjectID will validate a given OIDC subject ID. It ensures that the field
// consists of valid ASCII characters, is not empty and
// does not exceed max characters.
// As per http://openid.net/specs/openid-connect-core-1_0.html#IDToken
// For sub : It MUST NOT exceed 255 ASCII characters in length.
func SubjectID(fname, body string) error {
	idx := strings.IndexFunc(body, func(r rune) bool {
		return r < 32 || r > 126
	})
	if idx != -1 {
		// non-ASCII characters exist
		return validation.NewFieldError(fname, "must not contain non-ASCII characters")
	}

	// Checking length
	r := []rune(body)
	l := len(r)

	if l < 1 {
		return validation.NewFieldError(fname, "must not be empty")
	}
	if l > 255 {
		return validation.NewFieldError(fname, "cannot exceed 255 characters")
	}
	return nil
}
