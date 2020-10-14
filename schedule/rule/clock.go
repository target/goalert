package rule

import (
	"database/sql/driver"
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
)

// Clock represents wall-clock time. It is a duration since midnight.
type Clock time.Duration

// ParseClock will return a new Clock value given a value in the format of '15:04' or '15:04:05'.
// The resulting value will be truncated to the minute.
func ParseClock(value string) (Clock, error) {
	var h, m int
	var s float64
	n, err := fmt.Sscanf(value, "%d:%d:%f", &h, &m, &s)
	if n == 2 && errors.Is(err, io.ErrUnexpectedEOF) {
		err = nil
	}
	if err != nil {
		return 0, err
	}
	if n < 2 {
		return 0, errors.New("invalid time format")
	}
	if n == 3 && (s < 0 || s >= 60) {
		return 0, errors.New("invalid seconds value")
	}

	if h < 0 || h > 23 {
		return 0, errors.New("invalid hours value")
	}
	if m < 0 || m > 59 {
		return 0, errors.New("invalid minutes value")
	}

	return NewClock(h, m), nil
}

// String returns a string representation of the format '15:04'.
func (c Clock) String() string {
	return fmt.Sprintf("%02d:%02d", c.Hour(), c.Minute())
}

// NewClock returns a Clock value equal to the provided 24-hour value and minute.
func NewClock(hour, minute int) Clock {
	return Clock(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)
}

// Minute returns the minute of the Clock value.
func (c Clock) Minute() int {
	r := time.Duration(c) % time.Hour
	return int(r / time.Minute)
}

// Hour returns the hour of the Clock value.
func (c Clock) Hour() int {
	return int(time.Duration(c) / time.Hour)
}

// Format will format the clock value using the same format string
// used by time.Time.
func (c Clock) Format(layout string) string {
	return time.Date(0, 0, 0, c.Hour(), c.Minute(), 0, 0, time.UTC).Format(layout)
}

// Value implements the driver.Valuer interface.
func (c Clock) Value() (driver.Value, error) {
	return c.String(), nil
}

// Scan implements the sql.Scanner interface.
func (c *Clock) Scan(value interface{}) error {
	var parsed Clock
	var err error
	switch t := value.(type) {
	case []byte:
		parsed, err = ParseClock(string(t))
	case string:
		parsed, err = ParseClock(t)
	case time.Time:
		parsed = NewClock(
			t.Hour(),
			t.Minute(),
		)
	default:
		return errors.Errorf("could not scan unknown type %T as Clock", t)
	}
	if err != nil {
		return err
	}

	*c = parsed
	return nil
}
