package validate

import (
	"time"

	"github.com/target/goalert/validation"
)

// Duration will ensure a value is between min and max duration (inclusive).
// A FieldError is returned otherwise.
func Duration(fname string, val, min, max time.Duration) error {
	if val < min {
		return validation.NewFieldError(fname, "must not be below "+min.String())
	}
	if val > max {
		return validation.NewFieldError(fname, "must not be over "+max.String())
	}

	return nil
}
