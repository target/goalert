package validate

import (
	"github.com/target/goalert/validation"
)

// Many will take multiple input error values, filter out nils
// and flatten any nested MultiFieldErrors.
//
// If a given error is not a FieldError, or MultiFieldError it is returned immediately.
//
// If all errs are nil, nil is returned.
// If only one error is present, it is returned.
func Many(errs ...error) error {
	flat := make([]validation.FieldError, 0, len(errs))

	for _, e := range errs {
		switch err := e.(type) {
		case validation.MultiFieldError:
			flat = append(flat, err.FieldErrors()...)
		case validation.FieldError:
			flat = append(flat, err)
		case error:
			return e
		case nil:
		default:
			return e
		}
	}
	if len(flat) == 0 {
		return nil
	}
	if len(flat) == 1 {
		return flat[0]
	}
	return validation.NewMultiFieldError(flat)
}
