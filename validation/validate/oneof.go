package validate

import (
	"fmt"
	"strings"

	"github.com/target/goalert/validation"
)

// OneOf will check that value is one of the provided options.
func OneOf[T comparable](fname string, value T, options ...T) error {
	for _, o := range options {
		if o == value {
			return nil
		}
	}

	msg := []string{}
	for _, o := range options {
		msg = append(msg, fmt.Sprintf("%v", o))
	}

	return validation.NewFieldError(fname, "must be one of: "+strings.Join(msg, ", "))
}
