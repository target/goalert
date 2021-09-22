package validate

import (
	"strings"

	"github.com/target/goalert/validation"
)

// TwlioFromValue will validate a from value as either a phone number, or messaging service SID starting with 'MG'.
func TwilioFromValue(fname, value string) error {
	switch {
	case strings.HasPrefix(value, "+"):
		return Phone(fname, value)
	case strings.HasPrefix(value, "MG"):
		return ASCII(fname, value, 2, 64)
	}

	return validation.NewFieldError(fname, "must be a valid phone number or alphanumeric sender ID")
}
