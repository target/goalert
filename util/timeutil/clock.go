package timeutil

import (
	"database/sql/driver"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/pkg/errors"
)

// PrevClock returns the most recent instant the clock would read the provided value.
// The same time will be returned if it is already true.
//
// If DST comes into effect, the returned time will be the soonest instant it would
// have become true.
//
// For example: From midnight, the next 2:30AM if DST would cause the clock to jump
// from 2:00AM to 3:00AM, the returned time would be 3:00AM.
func PrevClock(t time.Time, c Clock) time.Time {

	return t
}

// FindNextOffsetChange will return the next timestamp (within 24 hours+limit)
// where the zone offset changes. The search is limited to times affected within
// the provided duration. Meaning if within that t + dur, there is a time that
// does not exist (spring forward) or repeats (fall back) the initial change time
// is returned.
//
// We do not support DST/zone changes > 24 hours, or any that last less than 48 hours :)
//
// If there is no change, zero time is returned
func FindNextOffsetChange(t time.Time, dur time.Duration) (at time.Time, changeBy Clock) {
	t = t.Truncate(time.Minute)
	dur = dur.Truncate(time.Minute)
	next := t.Add(dur + 24*time.Hour)

	_, oldOffset := t.Zone()
	_, newOffset := next.Zone()
	if oldOffset == newOffset {
		return time.Time{}, 0
	}

	if newOffset < oldOffset {
		// add fall-back amount to duration limit
		dur += (time.Duration(oldOffset-newOffset) * time.Second).Truncate(time.Minute)
	}
	next = t.Add(dur)
	_, newOffset = next.Zone()
	// change is not within search limit
	if oldOffset == newOffset {
		return time.Time{}, 0
	}

	// find when the zone offset changes
	mins := sort.Search(int(dur/time.Minute), func(min int) bool {
		_, n := t.Add(time.Duration(min) * time.Minute).Zone()
		return n == newOffset
	})

	return t.Add(time.Duration(mins) * time.Minute), Clock(time.Duration(newOffset-oldOffset) * time.Second)
}

// IsDST will return true if there is a DST change within 24-hours AFTER t.
//
// If so, the clock-time and amount of change is calculated.
func IsDST(t time.Time) (dst bool, at, change Clock) {

	next := t.Add(24 * time.Hour)
	_, oldOffset := t.Zone()
	_, newOffset := next.Zone()
	if oldOffset == newOffset {
		return false, 0, 0
	}

	mins := sort.Search(int(24*time.Hour/time.Minute), func(min int) bool {
		_, n := t.Add(time.Duration(min) * time.Minute).Zone()
		return n == newOffset
	})

	return true, NewClock(0, mins), Clock(time.Duration(newOffset-oldOffset) * time.Second)
}

// FirstOfDay will return the first timestamp where the time matches
// the clock value, or the first instant after, if it does not exist.
func (c Clock) FirstOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	t = time.Date(y, m, d, 0, 0, 0, 0, t.Location())

	isDST, dstAt, dstChange := IsDST(t)
	if !isDST || c < dstAt {
		// if we spring forward, DST won't happen yet
		// if we fall back, we'll land on the 'first' instance
		return t.Add(time.Duration(c))
	}
	// c >= dstAt

	if dstChange > 0 {
		if c < dstAt+dstChange {
			// e.g. 2:30AM when we go from 2AM->3AM, so return 3AM.
			return t.Add(time.Duration(dstAt))
		}
		// spring forward, so lose the amount of time
		return t.Add(time.Duration(c - dstChange))
	}

	// falls back and the target time is >= the fallback time
	// so add extra clock time
	return t.Add(time.Duration(c + -dstChange))
}

// LastOfDay will return the last timestamp where the time matches
// the clock value, or the first instant after, if it does not exist.
func (c Clock) LastOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	t = time.Date(y, m, d, 0, 0, 0, 0, t.Location())

	isDST, dstAt, dstChange := IsDST(t)
	if !isDST || (dstChange > 0 && c < dstAt) {
		return t.Add(time.Duration(c))
	}

	if dstChange > 0 {
		if c < dstAt+dstChange {
			// e.g. 2:30AM when we go from 2AM->3AM, so return 3AM.
			return t.Add(time.Duration(dstAt))
		}
		// >=dstAt so subtract the change
		return t.Add(time.Duration(c - dstChange))
	}

	dstRepeatAt := dstAt + dstChange
	if c < dstRepeatAt {
		return t.Add(time.Duration(c))
	}

	return t.Add(time.Duration(c + -dstChange))
}

// NextClock returns the next instant that the clock would read the provided value.
// The returned value is always in the future.
//
// If DST comes into effect, the returned time will be the soonest instant it would
// have become true.
//
// For example: From midnight, the next 2:30AM if DST would cause the clock to jump
// from 2:00AM to 3:00AM, the returned time would be 3:00AM.
func NextClock(t time.Time, c Clock) time.Time {
	h, m, _ := t.Clock()
	tClock := NewClock(h, m)

	diff := c - tClock
	if diff <= 0 {
		diff += Clock(24 * time.Hour)
	}

	at, change := FindNextOffsetChange(t, time.Duration(diff))
	if change == 0 {
		return t.Add(time.Duration(diff))
	}

	h, m, _ = at.Clock()

	endClock := NewClock(h, m)
	startClock := endClock - change
	if startClock <= c && c < endClock {
		// we end up in the middle of the non-existant span, so align to the start
		return t.Add(time.Duration(diff - c + startClock))
	}
	if startClock < endClock {
		return t.Add(time.Duration(diff - change))
	}

	// time falls back
	diff = c - tClock
	if diff <= 0 || c >= startClock {
		return t.Add(time.Duration(diff - change))
	}

	return t.Add(time.Duration(diff))
}

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

// Is returns true if t represents the same clock time.
func (c Clock) Is(t time.Time) bool {
	h, m, _ := t.Clock()
	return NewClock(h, m) == c
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
