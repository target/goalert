package validate

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/target/goalert/validation"
)

const upperLimit = 256

// SanitizeName will remove all invalid characters
// and return a valid name (as defined by ValidateName) or an empty string.
//
// It is used in cases where the input is not provided by a user, and should
// be used as-is if possible. An example would be importing a user profile from
// GitHub.
//
// If longer than upperLimit, the extra characters are dropped.
func SanitizeName(name string) string {
	// strip out anything that's not a letter or digit
	// and normalize spaces
	name = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, name)

	// trim leading/trailing spaces
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "  ", " ")

	r := []rune(name)
	if len(r) < 1 {
		return ""
	}
	if len(r) > upperLimit {
		return strings.TrimSpace(string(r[:upperLimit]))
	}

	return name
}

// Name will validate a unicode name to ensure it is between 1 and upperLimit runes,
// and only consists of printable unicode characters.
//
// If invalid, a FieldError with the given field name is returned.
func Name(fname, name string) error {
	r := []rune(name)
	l := len(r)
	if l < 1 {
		return validation.NewFieldError(fname, "must not be empty")
	}
	if l > upperLimit {
		return validation.NewFieldError(fname, fmt.Sprintf("cannot be more than %d characters", upperLimit))
	}

	idx := strings.IndexFunc(name, func(r rune) bool {
		if unicode.IsSpace(r) && r != ' ' {
			// whitespace other than space (chr 32)
			return true
		}
		return !unicode.IsPrint(r)
	})
	if idx != -1 {
		return validation.NewFieldError(fname, "can only contain printable characters")
	}

	if strings.TrimSpace(name) != name {
		return validation.NewFieldError(fname, "must not begin or end with space")
	}

	return nil
}
