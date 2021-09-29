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
		return TwilioSID(fname, "MG", value)
	}

	return validation.NewFieldError(fname, "must be a valid phone number or alphanumeric sender ID")
}

// TwilioSID will validate the format of a Twilio SID with the given prefix.
func TwilioSID(fname, prefix, value string) error {
	if !strings.HasPrefix(value, prefix) {
		return validation.NewFieldError(fname, "must start with "+prefix)
	}

	return ASCII(fname, value, len(prefix), 64)
}
