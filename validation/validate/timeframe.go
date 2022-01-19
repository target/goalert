package validate

import (
	"time"

	"github.com/target/goalert/validation"
)

func TimeFrame(fname string, since time.Time, until time.Time, timeframe time.Duration) error {
	if until.Sub(since) > timeframe {
		return validation.NewFieldError(fname, "since and until must not be more than "+timeframe.String() + " apart")
	}

	return nil
}