package validate

import (
	"github.com/target/goalert/validation"
	"regexp"
	"strings"
)

var labelKeyPrefixRx = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{2,63}(\.[a-z0-9][a-z0-9-]{2,63})*$`)

// LabelKey will validate a label key field to ensure it follows a particular format.
//
// A label key consists of a prefix (lowercase alphanumeric /w hyphens -- domain name rules) followed by
// a `/` and a suffix consisting of alphanumeric characters and hyphens.
// The entire key may not exceed 255 characters.
func LabelKey(fname, body string) error {
	// Checking length
	r := []rune(body)
	l := len(r)

	if l < 1 {
		return validation.NewFieldError(fname, "must not be empty")
	}
	if l > 255 {
		return validation.NewFieldError(fname, "cannot exceed 255 characters")
	}

	parts := strings.SplitN(body, "/", 2)
	if len(parts) != 2 {
		return validation.NewFieldError(fname, "prefix and suffix must be separated by `/`")
	}

	prefix := parts[0]
	suffix := parts[1]

	if len(prefix) < 3 {
		return validation.NewFieldError(fname, "prefix: must be at least 3 characters")
	}

	if len(suffix) == 0 {
		return validation.NewFieldError(fname, "suffix: must not be empty")
	}

	if (prefix[0] < 48 || prefix[0] > 57) && (prefix[0] < 97 || prefix[0] > 122) {
		return validation.NewFieldError(fname, "prefix: must begin with a lower-case letter or number")
	}

	idx := strings.IndexFunc(prefix, func(r rune) bool {
		if r == '.' || r == '-' {
			return false
		}
		if r >= 48 && r <= 57 { // numbers
			return false
		}
		if r >= 97 && r <= 122 { // lowercase letters
			return false
		}

		// anything else
		return true
	})
	if idx != -1 {
		return validation.NewFieldError(fname, "prefix: may only contain lowercase letters, numbers, hyphens, or periods")
	}

	if !labelKeyPrefixRx.MatchString(prefix) {
		return validation.NewFieldError(fname, "prefix: must follow domain name formatting")
	}

	type reasoner interface {
		Reason() string
	}
	err := LabelValue(fname, suffix)
	if err != nil {
		return validation.NewFieldError(fname, "suffix: "+err.(reasoner).Reason())
	}

	return nil
}
