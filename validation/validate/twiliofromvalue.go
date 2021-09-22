package validate

import (
	"github.com/target/goalert/validation"
)

// TwlioFromValue will validate a from value as either a phone number, or messenger SID
func TwilioFromValue(fname, value string) error {
	phoneErr := Phone(fname, value)
	sidErr := MessagingServiceSID(fname, value)
	asciiErr := ASCII(fname, value, 2, 64)

	if phoneErr != nil && sidErr != nil && asciiErr != nil {
		return validation.NewFieldError("From", "is not a valid phone number or alphanumeric sender ID.")
	}

	return nil
}