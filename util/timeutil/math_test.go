package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddClock(t *testing.T) {

	check := func(exp string, ts time.Time, c Clock) {
		t.Helper()
		res := AddClock(ts, c)
		assert.Equal(t, exp, res.String())
	}

	loc, err := time.LoadLocation("America/Chicago")
	require.NoError(t, err)

	check(
		"2018-03-11 01:00:00 -0600 CST",
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
		NewClock(1, 0),
	)
	check(
		"2018-03-11 03:00:00 -0500 CDT",
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
		NewClock(2, 0),
	)
	check(
		"2018-03-11 03:00:00 -0500 CDT",
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
		NewClock(2, 30),
	)

	check(
		"2017-01-03 12:00:00 -0600 CST",
		time.Date(2017, time.January, 1, 0, 0, 0, 0, loc),
		NewClock(60, 0),
	)

	check(
		"2020-11-01 01:30:00 -0500 CDT",
		time.Date(2020, 11, 2, 1, 30, 0, 0, loc),
		NewClock(-24, 0),
	)
}

func TestHoursBetween(t *testing.T) {
	check := func(exp int, a, b time.Time) {
		t.Helper()
		res := HoursBetween(a, b)
		assert.Equal(t, exp, res)
	}

	loc, err := time.LoadLocation("America/Chicago")
	require.NoError(t, err)
	// normal
	check(
		4,
		time.Date(2018, time.March, 9, 0, 0, 0, 0, loc),
		time.Date(2018, time.March, 9, 4, 0, 0, 0, loc),
	)

	// CST -> CDT
	check(
		4,
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
		time.Date(2018, time.March, 11, 4, 0, 0, 0, loc),
	)

	// CDT -> CST
	check(
		4,
		time.Date(2018, time.November, 4, 0, 0, 0, 0, loc),
		time.Date(2018, time.November, 4, 4, 0, 0, 0, loc),
	)

	check(
		49,
		time.Date(2020, 10, 30, 1, 30, 0, 0, loc),
		time.Date(2020, 11, 1, 2, 30, 0, 0, loc),
	)
}

func TestAddClock_MidnightDST(t *testing.T) {
	check := func(exp string, ts time.Time, c Clock) {
		t.Helper()
		res := AddClock(ts, c)
		assert.Equal(t, exp, res.String(), "AddClock(%s, %s)", ts, c)
	}

	// DST starts at midnight in Santiago on 2026-09-06 (00:00 -> 01:00)
	loc, err := time.LoadLocation("America/Santiago")
	require.NoError(t, err)

	// 00:30 on the 6th does not exist, so the first instant after is used
	check("2026-09-06 01:00:00 -0300 -03", time.Date(2026, time.September, 5, 0, 30, 0, 0, loc), NewClock(24, 0))
	check("2026-09-06 01:00:00 -0300 -03", time.Date(2026, time.September, 5, 23, 30, 0, 0, loc), NewClock(1, 0))
	check("2026-09-06 01:00:00 -0300 -03", time.Date(2026, time.September, 7, 0, 30, 0, 0, loc), NewClock(-24, 0))

	check("2026-09-06 08:00:00 -0300 -03", time.Date(2026, time.September, 5, 20, 0, 0, 0, loc), NewClock(12, 0))
	check("2026-09-06 08:00:00 -0300 -03", time.Date(2026, time.September, 7, 8, 0, 0, 0, loc), NewClock(-24, 0))
	check("2026-09-07 00:30:00 -0300 -03", time.Date(2026, time.September, 6, 1, 0, 0, 0, loc), NewClock(23, 30))
}
