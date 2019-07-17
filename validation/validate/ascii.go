package validate

import (
	"strconv"
	"unicode"

	"github.com/target/goalert/validation"
)

// ASCII will validate that the passed value is an ASCII string within the given length values.
func ASCII(fname, val string, minLen, maxLen int) error {
	l := len(val)
	if minLen > 1 && l < minLen {
		return validation.NewFieldError(fname, "must be at least "+strconv.Itoa(minLen)+" characters")
	} else if minLen == 1 && l < minLen {
		return validation.NewFieldError(fname, "must not be empty")
	}
	if l > maxLen {
		return validation.NewFieldError(fname, "cannot exceed "+strconv.Itoa(maxLen)+" characters")
	}

	for _, r := range val {
		if r >= 32 && r <= 126 {
			continue
		}
		if unicode.IsPrint(r) {
			return validation.NewFieldError(fname, "invalid character '"+string(r)+"'")
		}
		return validation.NewFieldError(fname, "invalid character 0x"+strconv.Itoa(int(r)))

	}

	return nil
}
