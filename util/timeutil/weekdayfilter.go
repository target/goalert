package timeutil

import (
	"bytes"
	"database/sql/driver"
	"encoding"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
)

type WeekdayFilter [7]byte

var _ encoding.TextMarshaler = WeekdayFilter{}
var _ encoding.TextUnmarshaler = &WeekdayFilter{}
var _ graphql.Marshaler = WeekdayFilter{}
var _ graphql.Unmarshaler = &WeekdayFilter{}

var (
	neverDays = WeekdayFilter([7]byte{})
	everyDay  = WeekdayFilter([7]byte{1, 1, 1, 1, 1, 1, 1})
)

func (f WeekdayFilter) IsNever() bool  { return f == neverDays }
func (f WeekdayFilter) IsAlways() bool { return f == everyDay }

// EveryDay returns a WeekdayFilter that is permanently active.
func EveryDay() WeekdayFilter { return everyDay }

// StartTime returns midnight of the day the filter became active, from the perspective of t.
//
// If the filter is active every day or no days, zero time is returned.
// If the current day is not active, zero time is returned.
func (f WeekdayFilter) StartTime(t time.Time) time.Time {
	w := int(t.Weekday())
	var days int
	for i := range f {
		day := (7 + (w - (i))) % 7

		// keep going until we find the first time we go from enabled to disabled
		if f[day] == 0 {
			break
		}
		days++
	}
	if days == 0 {
		return f.NextActive(t)
	}
	year, mon, day := t.Date()
	return time.Date(year, mon, day-days+1, 0, 0, 0, 0, t.Location())
}

// NextActive returns the next time, at midnight, from t that is active.
//
// If the filter is active every day or no days, zero time is returned.
// Otherwise the returned value will always be in the future.
func (f WeekdayFilter) NextActive(t time.Time) time.Time {
	w := int(t.Weekday())
	for i := range f {
		day := (w + (i + 1)) % 7

		if f[day] == 1 {
			return NextWeekday(t, time.Weekday(day))
		}
	}

	// no active days
	return time.Time{}
}

// NextInactive returns the next time, at midnight, from t that is no longer active.
//
// If the filter is active every day or no days, zero time is returned.
// Otherwise the returned value will always be in the future.
func (f WeekdayFilter) NextInactive(t time.Time) time.Time {
	w := int(t.Weekday())
	for i := range f {
		day := (w + (i + 1)) % 7

		if f[day] == 0 {
			return NextWeekday(t, time.Weekday(day))
		}
	}

	// no disabled days
	return time.Time{}
}

// Day will return true if the given weekday is enabled.
func (f WeekdayFilter) Day(d time.Weekday) bool {
	if d < 0 {
		d += 7
	}
	return f[int(d)%7] == 1
}

// SetDay will update the filter for the given weekday.
func (f *WeekdayFilter) SetDay(d time.Weekday, enabled bool) {
	if enabled {
		f[int(d)] = 1
	} else {
		f[int(d)] = 0
	}
}

// DaysUntil will give the number of days until
// a matching day from the given weekday. -1 is returned
// if no days match.
func (f WeekdayFilter) DaysUntil(d time.Weekday, enabled bool) int {
	if enabled && f == neverDays {
		return -1
	}
	if !enabled && f == everyDay {
		return -1
	}
	var val byte
	if enabled {
		val = 1
	}
	idx := bytes.IndexByte(f[d:], val)
	if idx > -1 {
		return idx
	}

	idx = bytes.IndexByte(f[:], val)
	return 7 - int(d) + idx
}

// DaysSince will give the number of days since
// an enabled day from the given weekday. -1 is returned
// if all days are disabled.
func (f WeekdayFilter) DaysSince(d time.Weekday, enabled bool) int {
	if enabled && f == neverDays {
		return -1
	}
	if !enabled && f == everyDay {
		return -1
	}

	var val byte
	if enabled {
		val = 1
	}
	idx := bytes.LastIndexByte(f[:d+1], val)
	if idx > -1 {
		return int(d) - idx
	}

	idx = bytes.LastIndexByte(f[d+1:], val)
	return 6 - idx
}

func (f WeekdayFilter) MarshalGQL(w io.Writer) {
	res := make([]bool, 7)
	for i := range f {
		res[i] = f[i] == 1
	}
	graphql.MarshalAny(res).MarshalGQL(w)
}

func (f *WeekdayFilter) UnmarshalGQL(v interface{}) error {
	slice, ok := v.([]interface{})
	if !ok {
		return validation.NewFieldError("weekdayFilter", "must be an array")
	}
	if len(slice) != 7 {
		return validation.NewFieldError("weekdayFilter", fmt.Sprintf("expected 7 items; got %d", len(slice)))
	}
	for i, v := range slice {
		b, ok := v.(bool)
		if !ok {
			return validation.NewFieldError(fmt.Sprintf("weekdayFilter[%d]", i), fmt.Sprintf("expected true or false; got %T", v))
		}
		if b {
			f[i] = 1
		} else {
			f[i] = 0
		}
	}

	return nil
}

func (f WeekdayFilter) MarshalText() ([]byte, error) {
	res := make([]byte, 7)
	for i, v := range f {
		if v == 0 {
			res[i] = '0'
			continue
		}

		res[i] = '1'
	}

	return res, nil
}
func (f *WeekdayFilter) UnmarshalText(data []byte) error {
	s := string(data)
	if s == "" {
		*f = neverDays
		return nil
	}

	if len(s) != 7 {
		return fmt.Errorf("invalid length: expected 7; got %d", len(s))
	}

	for i, r := range s {
		switch r {
		case '0':
			f[i] = 0
		case '1':
			f[i] = 1
		default:
			return fmt.Errorf("invalid character at position %d: expected 0 or 1 but got %c", i, r)
		}
	}

	return nil
}

// String returns a string representation of the WeekdayFilter.
func (f WeekdayFilter) String() string {
	switch f {
	case WeekdayFilter{1, 0, 0, 0, 0, 0, 1}:
		return "weekends"
	case neverDays:
		return "never"
	case everyDay:
		return "every day"
	case WeekdayFilter{0, 1, 1, 1, 1, 1, 0}:
		return "M-F"
	case WeekdayFilter{0, 1, 1, 1, 1, 1, 1}:
		return "M-F and Sat"
	case WeekdayFilter{1, 1, 1, 1, 1, 1, 0}:
		return "M-F and Sun"
	}
	var days []string
	var chain []time.Weekday
	flushChain := func() {
		if len(chain) < 3 {
			for _, wd := range chain {
				days = append(days, wd.String()[:3])
			}
			chain = chain[:0]
			return
		}

		days = append(days, chain[0].String()[:3]+"-"+chain[len(chain)-1].String()[:3])
		chain = chain[:0]
	}
	for d, act := range f {
		if act == 1 {
			chain = append(chain, time.Weekday(d))
			continue
		}
		flushChain()
	}
	flushChain()

	return strings.Join(days, ",")
}

// Value converts the WeekdayFilter to a DB array of bool.
func (f WeekdayFilter) Value() (driver.Value, error) {
	return sqlutil.BoolArray{
		f[time.Sunday] != 0,
		f[time.Monday] != 0,
		f[time.Tuesday] != 0,
		f[time.Wednesday] != 0,
		f[time.Thursday] != 0,
		f[time.Friday] != 0,
		f[time.Saturday] != 0,
	}.Value()
}

// Scan scans the WeekdayFilter from a DB array of bool.
func (f *WeekdayFilter) Scan(src interface{}) error {
	var b sqlutil.BoolArray
	err := b.Scan(src)
	if err != nil {
		return err
	}
	for i := range f {
		if i < len(b) && b[i] {
			f[i] = 1
		} else {
			f[i] = 0
		}
	}

	return nil
}
