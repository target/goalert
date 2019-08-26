package rule

import (
	"bytes"
	"database/sql/driver"
	"strings"
	"time"

	"github.com/target/goalert/util/sqlutil"
)

type WeekdayFilter [7]byte

var (
	neverDays = WeekdayFilter([7]byte{})
	everyDay  = WeekdayFilter([7]byte{1, 1, 1, 1, 1, 1, 1})
)

// Day will return true if the given weekday is enabled.
func (f WeekdayFilter) Day(d time.Weekday) bool {
	return f[int(d)] == 1
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
