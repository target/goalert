package validate

import (
	"strings"

	"github.com/target/goalert/validation"
)

// MessagingServiceSID will validate an Messaging Service SID, returning a FieldError if invalid.
func MessagingServiceSID(fname, value string) error {
	if !strings.HasPrefix(value, "MG") {
		return validation.NewFieldError(fname, "must begin with MG")
	}

	return nil
}
