package validate

import (
	"github.com/target/goalert/validation"
	"unicode"
)

// Search will validate a search body. It ensures that the field 
// consists of valid unicode code-points, and does not exceed max of 255 characters.
// If body is empty, the input is considered valid.
func Search(fname, body string) error {
	if body == "" {
		return nil
	}

	r := []rune(body)
	l := len(r)

	if l == 0 {
		return nil
	}
	if l > 255 {
		return validation.NewFieldError(fname, "cannot exceed 255 characters")
	}

	for _, c := range r {
		if !unicode.IsPrint(c) && c != '\t' && c != '\n' {
			return validation.NewFieldError(fname, "only printable characters allowed")
		}
	}

	return nil
}
