package validate

import (
	"github.com/target/goalert/validation"
)

// TwlioFromValue will validate a from value as either a phone number, or messenger SID
func TwilioFromValue(fname, value string) error {
	phoneErr := Phone(fname, value)
	sidErr := TwilioMessageSID(fname, value)

	if phoneErr != nil && sidErr != nil {
		return validation.NewFieldError("From", "is not a valid phone number or alphanumeric sender ID.")
	}

	return nil
}