package timeutil

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeekdayFilter_StartTime(t *testing.T) {
	var loc *time.Location
	setLocation := func(newLoc *time.Location, err error) {
		t.Helper()
		loc = newLoc
		require.NoError(t, err)
	}
	check := func(expTime, start string, filter WeekdayFilter) {
		t.Helper()

		ts, err := time.ParseInLocation("2006-01-02 15:04:05.999999999 -0700 MST", start, loc)
		if err != nil {
			// workaround for being unable to parse "+1030" as "MST"
			parts := strings.Split(start, " ")
			ts, err = time.ParseInLocation("2006-01-02 15:04:05.999999999 -0700", strings.Join(parts[:3], " "), loc)
		}
		require.NoError(t, err)

		res := filter.StartTime(ts)

		assert.Equalf(t, expTime, res.String(),
			"start time %s from %s (%s, filter=%v)", res.String(), start, loc.String(), filter,
		)

	}

	setLocation(time.LoadLocation("America/Chicago"))
	check(
		"2020-10-28 00:00:00 -0500 CDT",
		"2020-10-30 01:00:00 -0500 CDT",
		WeekdayFilter{1, 1, 0, 1, 1, 1, 1},
	)
	check(
		"2020-10-28 00:00:00 -0500 CDT",
		"2020-11-01 01:00:00 -0500 CDT",
		WeekdayFilter{1, 1, 0, 1, 1, 1, 1},
	)
	check(
		"2020-11-04 00:00:00 -0500 CDT",
		"2020-11-03 01:00:00 -0500 CDT",
		WeekdayFilter{1, 1, 0, 1, 1, 1, 1},
	)
}

func TestWeekdayFilter_NextInactive(t *testing.T) {
	var loc *time.Location
	setLocation := func(newLoc *time.Location, err error) {
		t.Helper()
		loc = newLoc
		require.NoError(t, err)
	}
	check := func(expTime, start string, filter WeekdayFilter) {
		t.Helper()

		ts, err := time.ParseInLocation("2006-01-02 15:04:05.999999999 -0700 MST", start, loc)
		if err != nil {
			// workaround for being unable to parse "+1030" as "MST"
			parts := strings.Split(start, " ")
			ts, err = time.ParseInLocation("2006-01-02 15:04:05.999999999 -0700", strings.Join(parts[:3], " "), loc)
		}
		require.NoError(t, err)

		res := filter.NextInactive(ts)

		assert.Equalf(t, expTime, res.String(),
			"next start of inactive day %s from %s (%s, filter=%v)", res.String(), start, loc.String(), filter,
		)

	}

	setLocation(time.UTC, nil)
	check(
		"2020-11-03 00:00:00 -0500 CDT",
		"2020-11-01 01:00:00 -0500 CDT",
		WeekdayFilter{1, 1, 0, 1, 1, 1, 1},
	)
	check(
		"2020-11-10 00:00:00 -0500 CDT",
		"2020-11-03 00:00:00 -0500 CDT",
		WeekdayFilter{1, 1, 0, 1, 1, 1, 1},
	)
	check(
		"2020-11-10 00:00:00 -0500 CDT",
		"2020-11-04 01:00:00 -0500 CDT",
		WeekdayFilter{1, 1, 0, 1, 1, 1, 1},
	)
}

func TestWeekdayFilter_String(t *testing.T) {
	check := func(f WeekdayFilter, exp string) {
		var name strings.Builder
		for i := range f {
			if f[i] == 1 {
				name.WriteString(time.Weekday(i).String()[:1])
			} else {
				name.WriteByte('_')
			}
		}
		t.Run(name.String(), func(t *testing.T) {
			res := f.String()
			if res != exp {
				t.Errorf("got '%s'; want '%s'", res, exp)
			}
		})
	}

	check(everyDay, "every day")
	check(neverDays, "never")
	check(WeekdayFilter{1, 0, 0, 0, 0, 0, 1}, "weekends")
	check(WeekdayFilter{0, 1, 1, 1, 1, 1, 0}, "M-F")
	check(WeekdayFilter{1, 1, 0, 0, 0, 1, 0}, "Sun,Mon,Fri")
}

func TestWeekdayFilter_DaysUntil(t *testing.T) {
	check := func(f WeekdayFilter, in time.Weekday, e bool, exp int) {
		t.Run(fmt.Sprintf("%s/From%s-%t", f, in, e), func(t *testing.T) {
			res := f.DaysUntil(in, e)
			if res != exp {
				t.Errorf("got %d; want %d", res, exp)
			}
		})
	}

	check(everyDay, time.Monday, true, 0)
	check(neverDays, time.Monday, true, -1)
	check(WeekdayFilter{1, 0, 0, 0, 0, 0, 0}, time.Monday, true, 6)
	check(WeekdayFilter{0, 1, 0, 0, 0, 0, 0}, time.Monday, true, 0)
	check(WeekdayFilter{0, 0, 1, 0, 0, 0, 0}, time.Monday, true, 1)
	check(WeekdayFilter{1, 0, 1, 0, 0, 1, 0}, time.Monday, true, 1)

	check(everyDay, time.Monday, false, -1)
	check(neverDays, time.Monday, false, 0)
	check(WeekdayFilter{1, 0, 0, 0, 0, 0, 0}, time.Monday, false, 0)
	check(WeekdayFilter{0, 1, 0, 0, 0, 0, 0}, time.Monday, false, 1)
	check(WeekdayFilter{0, 0, 1, 0, 0, 0, 0}, time.Monday, false, 0)
	check(WeekdayFilter{1, 0, 1, 0, 0, 1, 0}, time.Monday, false, 0)
	check(WeekdayFilter{0, 1, 1, 1, 1, 1, 1}, time.Monday, false, 6)
}

func TestWeekdayFilter_DaysSince(t *testing.T) {
	check := func(f WeekdayFilter, in time.Weekday, e bool, exp int) {
		t.Run(fmt.Sprintf("%s/From%s-%t", f, in, e), func(t *testing.T) {
			res := f.DaysSince(in, e)
			if res != exp {
				t.Errorf("got %d; want %d", res, exp)
			}
		})
	}

	check(everyDay, time.Monday, true, 0)
	check(neverDays, time.Monday, true, -1)
	check(WeekdayFilter{1, 0, 0, 0, 0, 0, 0}, time.Monday, true, 1)
	check(WeekdayFilter{0, 1, 0, 0, 0, 0, 0}, time.Monday, true, 0)
	check(WeekdayFilter{0, 0, 1, 0, 0, 0, 0}, time.Monday, true, 6)
	check(WeekdayFilter{1, 0, 1, 0, 0, 1, 0}, time.Monday, true, 1)
	check(WeekdayFilter{0, 0, 1, 0, 0, 1, 0}, time.Monday, true, 3)
	check(WeekdayFilter{0, 0, 1, 0, 0, 1, 1}, time.Monday, true, 2)

	check(everyDay, time.Monday, false, -1)
	check(neverDays, time.Monday, false, 0)
	check(WeekdayFilter{1, 0, 0, 0, 0, 0, 0}, time.Monday, false, 0)
	check(WeekdayFilter{0, 1, 0, 0, 0, 0, 0}, time.Monday, false, 1)
	check(WeekdayFilter{0, 0, 1, 0, 0, 0, 0}, time.Monday, false, 0)
	check(WeekdayFilter{1, 0, 1, 0, 0, 1, 0}, time.Monday, false, 0)
	check(WeekdayFilter{0, 0, 1, 0, 0, 1, 0}, time.Monday, false, 0)
	check(WeekdayFilter{0, 0, 1, 0, 0, 1, 1}, time.Monday, false, 0)
	check(WeekdayFilter{1, 1, 1, 0, 0, 1, 1}, time.Monday, false, 4)

}

func TestWeekdayFilter_MidnightDST(t *testing.T) {
	// DST starts at midnight in Santiago on Sunday 2026-09-06 (00:00 -> 01:00),
	// so the day starts at 01:00 -03.
	loc, err := time.LoadLocation("America/Santiago")
	require.NoError(t, err)

	fri := time.Date(2026, time.September, 4, 10, 0, 0, 0, loc)
	sat := time.Date(2026, time.September, 5, 23, 30, 0, 0, loc)
	sun := time.Date(2026, time.September, 6, 10, 0, 0, 0, loc)
	mon := time.Date(2026, time.September, 7, 10, 0, 0, 0, loc)

	sunOnly := WeekdayFilter{1, 0, 0, 0, 0, 0, 0}
	noSun := WeekdayFilter{0, 1, 1, 1, 1, 1, 1}
	noSat := WeekdayFilter{1, 1, 1, 1, 1, 1, 0}

	const sunStart = "2026-09-06 01:00:00 -0300 -03"
	assert.Equal(t, sunStart, sunOnly.NextActive(fri).String(), "NextActive from Friday")
	assert.Equal(t, sunStart, sunOnly.NextActive(sat).String(), "NextActive from Saturday")
	assert.Equal(t, sunStart, noSun.NextInactive(fri).String(), "NextInactive from Friday")
	assert.Equal(t, sunStart, noSun.NextInactive(sat).String(), "NextInactive from Saturday")

	assert.Equal(t, sunStart, sunOnly.StartTime(sun).String(), "StartTime on Sunday")
	assert.Equal(t, sunStart, noSat.StartTime(sun).String(), "StartTime on Sunday (start of streak)")
	assert.Equal(t, sunStart, noSat.StartTime(mon).String(), "StartTime on Monday (streak started Sunday)")
	assert.Equal(t, sunStart, sunOnly.StartTime(sat).String(), "StartTime on inactive Saturday (next active)")

	// the following day has a normal midnight
	const monStart = "2026-09-07 00:00:00 -0300 -03"
	assert.Equal(t, monStart, noSun.NextActive(sun).String(), "NextActive from Sunday")
	assert.Equal(t, monStart, sunOnly.NextInactive(sun).String(), "NextInactive from Sunday")
}
