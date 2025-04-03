package validate

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/target/goalert/validation"
)

// SanitizeText will sanitize a text body so that it passes the Text validation.
func SanitizeText(body string, maxLen int) string {
	body = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) || r == '\t' || r == '\n' {
			return r
		}
		if unicode.IsSpace(r) {
			return ' '
		}
		return -1
	}, body)

	// remove trailing space from lines
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	body = strings.Join(lines, "\n")

	// strip multiple newlines (more than 2)
	body = strings.ReplaceAll(body, "\n\n\n", "\n\n")
	body = strings.TrimSpace(body)

	r := []rune(body)
	if maxLen > 0 && len(r) > maxLen {
		// truncate the message to fit maxLen if needed
		return string(r[:maxLen-1]) + "…"
	}
	return body
}

// Text will validate a text body. It ensures that the field
// consists of valid unicode code-points, has at least min characters and
// does not exceed max characters, and that it doesn't begin or end with space.
//
// If body is empty, the input is considered valid, regardless of min value.
func Text(fname, body string, min, max int) error {
	if body == "" {
		return nil
	}

	return RequiredText(fname, body, min, max)
}

// RequiredText works like Text, but does not allow it to be blank, unless min is set to 0.
func RequiredText(fname, body string, min, max int) error {
	r := []rune(body)
	l := len(r)

	if l == 0 && min == 0 {
		return nil
	}

	if min > 1 && l < min {
		return validation.NewFieldError(fname, "must be at least "+strconv.Itoa(min)+" characters")
	} else if min == 1 && l < min {
		return validation.NewFieldError(fname, "must not be empty")
	}
	if l > max {
		return validation.NewFieldError(fname, "cannot exceed "+strconv.Itoa(max)+" characters")
	}

	for _, c := range r {
		if !unicode.IsPrint(c) && c != '\t' && c != '\n' {
			return validation.NewFieldError(fname, "only printable characters allowed")
		}
	}
	if unicode.IsSpace(r[0]) {
		return validation.NewFieldError(fname, "cannot begin with a space")
	}
	if unicode.IsSpace(r[l-1]) {
		return validation.NewFieldError(fname, "cannot end with a space")
	}
	return nil
}
