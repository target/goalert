package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNextWeekday(t *testing.T) {
	check := func(exp string, ts time.Time, w time.Weekday) {
		t.Helper()
		res := NextWeekday(ts, w)
		assert.Equal(t, exp, res.String(), "NextWeekday(%s, %s)", ts, w)
		assert.True(t, res.After(ts), "NextWeekday(%s, %s) should be in the future", ts, w)
	}

	// 2026-07-04 is a Saturday
	chicago := mustLoadLocation(t, "America/Chicago")
	check("2026-07-06 00:00:00 -0500 CDT", time.Date(2026, time.July, 4, 8, 0, 0, 0, chicago), time.Monday)
	check("2026-07-11 00:00:00 -0500 CDT", time.Date(2026, time.July, 4, 8, 0, 0, 0, chicago), time.Saturday)
	check("2026-07-11 00:00:00 -0500 CDT", time.Date(2026, time.July, 4, 0, 0, 0, 0, chicago), time.Saturday)

	// DST starts at midnight in Santiago on Sunday 2026-09-06 (00:00 -> 01:00)
	santiago := mustLoadLocation(t, "America/Santiago")
	check("2026-09-06 01:00:00 -0300 -03", time.Date(2026, time.September, 4, 10, 0, 0, 0, santiago), time.Sunday)
	check("2026-09-06 01:00:00 -0300 -03", time.Date(2026, time.September, 5, 23, 30, 0, 0, santiago), time.Sunday)
	check("2026-09-07 00:00:00 -0300 -03", time.Date(2026, time.September, 6, 1, 0, 0, 0, santiago), time.Monday)
	check("2026-09-13 00:00:00 -0300 -03", time.Date(2026, time.September, 6, 1, 0, 0, 0, santiago), time.Sunday)

	// DST starts at midnight in Havana on Sunday 2026-03-08 (00:00 -> 01:00)
	havana := mustLoadLocation(t, "America/Havana")
	check("2026-03-08 01:00:00 -0400 CDT", time.Date(2026, time.March, 6, 10, 0, 0, 0, havana), time.Sunday)
	check("2026-03-08 01:00:00 -0400 CDT", time.Date(2026, time.March, 7, 23, 59, 0, 0, havana), time.Sunday)
}

func TestPrevWeekday(t *testing.T) {
	check := func(exp string, ts time.Time, w time.Weekday) {
		t.Helper()
		res := PrevWeekday(ts, w)
		assert.Equal(t, exp, res.String(), "PrevWeekday(%s, %s)", ts, w)
		assert.False(t, res.After(ts), "PrevWeekday(%s, %s) should not be in the future", ts, w)
	}

	// 2026-07-04 is a Saturday
	chicago := mustLoadLocation(t, "America/Chicago")
	check("2026-06-29 00:00:00 -0500 CDT", time.Date(2026, time.July, 4, 8, 0, 0, 0, chicago), time.Monday)
	check("2026-07-04 08:00:00 -0500 CDT", time.Date(2026, time.July, 4, 8, 0, 0, 0, chicago), time.Saturday) // same day is returned as-is

	// DST starts at midnight in Santiago on Sunday 2026-09-06 (00:00 -> 01:00)
	santiago := mustLoadLocation(t, "America/Santiago")
	check("2026-09-06 01:00:00 -0300 -03", time.Date(2026, time.September, 8, 10, 0, 0, 0, santiago), time.Sunday)
	check("2026-09-06 01:00:00 -0300 -03", time.Date(2026, time.September, 7, 0, 0, 0, 0, santiago), time.Sunday)
	check("2026-09-05 00:00:00 -0400 -04", time.Date(2026, time.September, 6, 1, 0, 0, 0, santiago), time.Saturday)

	// DST starts at midnight in Havana on Sunday 2026-03-08 (00:00 -> 01:00)
	havana := mustLoadLocation(t, "America/Havana")
	check("2026-03-08 01:00:00 -0400 CDT", time.Date(2026, time.March, 9, 10, 0, 0, 0, havana), time.Sunday)
}
