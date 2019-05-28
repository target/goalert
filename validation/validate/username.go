package validate

import "github.com/target/goalert/validation"

// UserName will validate a username to ensure it is between 3 and 24 characters,
// and only contains lower-case ASCII letters and numbers.
func UserName(fname, name string) error {
	b := []byte(name)
	l := len(b)
	if l < 3 {
		return validation.NewFieldError(fname, "must be at least 3 characters")
	}
	if l > 24 {
		return validation.NewFieldError(fname, "cannot be more than 24 characters")
	}

	for _, c := range name {
		if c >= 'a' && c <= 'z' {
			continue
		}
		if c >= '0' && c <= '9' {
			continue
		}

		return validation.NewFieldError(fname, "can only contain lower-case letters and digits")
	}

	return nil
}
