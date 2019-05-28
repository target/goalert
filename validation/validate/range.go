package validate

import (
	"github.com/target/goalert/validation"
	"strconv"
)

// Range will ensure a value is between min and max (inclusive).
// A FieldError is returned otherwise.
func Range(fname string, val, min, max int) error {
	if min == 0 && val < 0 {
		return validation.NewFieldError(fname, "must not be negative")
	}
	if val < min {
		return validation.NewFieldError(fname, "must not be below "+strconv.Itoa(min))
	}
	if val > max {
		return validation.NewFieldError(fname, "must not be over "+strconv.Itoa(max))
	}

	return nil
}
