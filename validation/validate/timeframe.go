package validate

import (
	"time"

	"github.com/target/goalert/validation"
)

func TimeFrame(fname string, since time.Time, until time.Time, timeFrame time.Duration) error {
	if until.Sub(since) > timeFrame {
		return validation.NewFieldError(fname, "must not be more than "+timeFrame.String() + " apart")
	}

	return nil
}