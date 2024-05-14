package validate

import (
	"strconv"

	"github.com/target/goalert/validation"
)

// Len will ensure a slice has at least min and at most max items.
func Len[T any](fname string, val []T, min, max int) error {
	if len(val) < min {
		return validation.NewFieldError(fname, "must have at least "+strconv.Itoa(min)+" items")
	}
	if len(val) > max {
		return validation.NewFieldError(fname, "must have at most "+strconv.Itoa(max)+" items")
	}
	return nil
}

// MapLen will ensure a map has at least min and at most max items.
func MapLen[K comparable, V any](fname string, val map[K]V, min, max int) error {
	if len(val) < min {
		return validation.NewFieldError(fname, "must have at least "+strconv.Itoa(min)+" items")
	}
	if len(val) > max {
		return validation.NewFieldError(fname, "must have at most "+strconv.Itoa(max)+" items")
	}
	return nil
}

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
