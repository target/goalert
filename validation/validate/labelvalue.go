package validate

import (
	"github.com/target/goalert/validation"
	"strings"
	"unicode"
)

// LabelValue will validate a label value field to ensure it consists of only printable characters as defined by Go.
// https://golang.org/pkg/unicode/#IsPrint
// It must be between 3 and 255 characters.
// If invalid, a FieldError with the given field name is returned.
func LabelValue(fname, body string) error {
	r := []rune(body)
	l := len(r)

	if l == 0 {
		return nil
	}

	if l < 3 {
		return validation.NewFieldError(fname, "must be at least 3 characters")
	}

	if l > 255 {
		return validation.NewFieldError(fname, "cannot exceed 255 characters")
	}

	if strings.TrimSpace(body) != body {
		return validation.NewFieldError(fname, "must not begin or end with a space")
	}
	if strings.Contains(body, "  ") {
		return validation.NewFieldError(fname, "must not contain double spaces")
	}

	for _, i := range r {
		if !unicode.IsPrint(i) {
			return validation.NewFieldError(fname, "must only contain printable characters")
		}
	}
	return nil
}
