package validate

import (
	"net/url"

	"github.com/target/goalert/validation"
)

// URI will validate a URI, returning a FieldError if invalid.
func URI(fname, value string) error {
	if _, err := url.ParseRequestURI(value); err != nil {
		return validation.NewFieldError(fname, "must be a valid uri: "+err.Error())
	}
	return nil
}
