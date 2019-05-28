package validate

import (
	"fmt"
	"github.com/target/goalert/validation"
	"strings"
)

// OneOf will check that value is one of the provided options.
func OneOf(fname string, value interface{}, options ...interface{}) error {
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
