package validate

import (
	"fmt"
	"github.com/target/goalert/validation"
)

func ArrayLength(fname string, array []int, max int) error {
	if len(array) > max {
		return validation.NewFieldError(fname, fmt.Sprintf("cannot contain more than %d", max))
	}
	return nil
}
