package validate

import (
	"strings"

	"github.com/target/goalert/validation"
)

// TwilioMessageSID will validate an Message SID, returning a FieldError if invalid.
func TwilioMessageSID(fname, value string) error {

	if !strings.HasPrefix(value, "MG") {
		return validation.NewFieldError(fname, "must begin with MG")
	}

	return nil
}
