package validate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nyaruka/phonenumbers"
	"github.com/target/goalert/validation"
)

var phoneRx = regexp.MustCompile(`^\+\d{1,15}$`)

// Phone will validate a phone number, returning a FieldError
// if invalid.
func Phone(fname, phone string) error {
	if !strings.HasPrefix(phone, "+") {
		return validation.NewFieldError(fname, "must contain country code")
	}
	if len(phone) < 2 {
		return validation.NewFieldError(fname, "must contain 1 or more digits")
	}
	if len(phone) > 16 {
		return validation.NewFieldError(fname, "must contain no more than 15 digits")
	}
	if !phoneRx.MatchString(phone) {
		return validation.NewFieldError(fname, "must only contain digits")
	}

	p, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return validation.NewFieldError(fname, fmt.Sprintf("must be a valid number: %s", err.Error()))
	}

	if !phonenumbers.IsValidNumber(p) {
		return validation.NewFieldError(fname, "must be a valid number")
	}
	return nil
}
