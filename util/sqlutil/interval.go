package sqlutil

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Interval = pgtype.Interval

// IntervalMicro returns a new Interval with the given duration in microseconds.
func IntervalMicro(dur time.Duration) Interval {
	return pgtype.Interval{Microseconds: int64(dur / time.Microsecond), Valid: true}
}
