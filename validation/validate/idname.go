package validate

import (
	"github.com/target/goalert/validation"
	"regexp"
)

var idRx = regexp.MustCompile(`^[a-zA-Z0-9 \-_']+$`)

// IDName will validate a ASCII name/identifier to ensure it is between 2 and 64 characters,
// starts with a letter, contains only letters, numbers, and spaces `-`, `_` or `'`.
//
// If invalid, a FieldError with the given field name is returned.
func IDName(fname, name string) error {
	b := []byte(name)
	l := len(b)
	if l < 2 {
		return validation.NewFieldError(fname, "must be at least 2 characters")
	}
	if l > 64 {
		return validation.NewFieldError(fname, "cannot be more than 64 characters")
	}

	if (b[0] < 'a' || b[0] > 'z') && (b[0] < 'A' || b[0] > 'Z') {
		return validation.NewFieldError(fname, "must begin with a letter")
	}

	if !idRx.Match(b) {
		return validation.NewFieldError(fname, "can only contain letters, digits, hyphens, underscores, apostrophe and space")
	}

	if b[l-1] == ' ' {
		return validation.NewFieldError(fname, "must not end with space")
	}

	return nil
}
