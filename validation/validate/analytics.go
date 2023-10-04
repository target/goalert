package validate

import (
	"strings"

	"github.com/target/goalert/validation"
)

// MeasurementID will validate the format of a Google Analytics 4 MeasurementID.
func MeasurementID(fname, value string) error {
	if !strings.HasPrefix(value, "G-") {
		return validation.NewFieldError(fname, "must start with G-")
	}

	return Text(fname, value, 0, 50)
}
