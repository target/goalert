package validate

import (
	"github.com/target/goalert/validation"
	"net/mail"
	"strings"
)

// SanitizeEmail will try to parse the email field and then return lower-case address portion or an empty string if parse failed.
func SanitizeEmail(email string) string {
	m, err := mail.ParseAddress(email)
	if err != nil {
		return ""
	}
	return strings.ToLower(m.Address)
}

// Email will validate an email address, returning a FieldError
// if invalid. Both named and un-named addresses are valid.
func Email(fname, email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return validation.NewFieldError(fname, "must be a valid email: "+err.Error())
	}
	return nil
}
