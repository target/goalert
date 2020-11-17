package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLastOfDay(t *testing.T) {
	check := func(expTs string, c Clock, ts time.Time) {
		t.Helper()
		res := c.LastOfDay(ts)
		assert.Equal(t, expTs, res.String())
	}

	loc, err := time.LoadLocation("America/Chicago")
	require.NoError(t, err)
	// normal
	check(
		"2018-03-09 01:00:00 -0600 CST",
		NewClock(1, 0),
		time.Date(2018, time.March, 9, 0, 0, 0, 0, loc),
	)

	// CST -> CDT
	check(
		"2018-03-11 01:00:00 -0600 CST",
		NewClock(1, 0),
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
	)
	check(
		"2018-03-11 01:30:00 -0600 CST",
		NewClock(1, 30),
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
	)
	check(
		"2018-03-11 03:00:00 -0500 CDT",
		NewClock(2, 0),
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
	)
	check(
		"2018-03-11 03:00:00 -0500 CDT",
		NewClock(2, 30),
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
	)
	check(
		"2018-03-11 03:00:00 -0500 CDT",
		NewClock(3, 0),
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
	)
	check(
		"2018-03-11 03:30:00 -0500 CDT",
		NewClock(3, 30),
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
	)

	// CDT -> CST
	check(
		"2018-11-04 00:30:00 -0500 CDT",
		NewClock(0, 30),
		time.Date(2018, time.November, 4, 0, 0, 0, 0, loc),
	)
	check(
		"2018-11-04 01:00:00 -0600 CST",
		NewClock(1, 0),
		time.Date(2018, time.November, 4, 0, 0, 0, 0, loc),
	)
	check(
		"2018-11-04 01:30:00 -0600 CST",
		NewClock(1, 30),
		time.Date(2018, time.November, 4, 0, 0, 0, 0, loc),
	)
	check(
		"2018-11-04 02:00:00 -0600 CST",
		NewClock(2, 0),
		time.Date(2018, time.November, 4, 0, 0, 0, 0, loc),
	)
	check(
		"2018-11-04 02:30:00 -0600 CST",
		NewClock(2, 30),
		time.Date(2018, time.November, 4, 0, 0, 0, 0, loc),
	)
}

func TestFirstOfDay(t *testing.T) {
	check := func(expTs string, c Clock, ts time.Time) {
		t.Helper()
		res := c.FirstOfDay(ts)
		assert.Equal(t, expTs, res.String())
	}

	loc, err := time.LoadLocation("America/Chicago")
	require.NoError(t, err)
	// normal
	check(
		"2018-03-09 01:00:00 -0600 CST",
		NewClock(1, 0),
		time.Date(2018, time.March, 9, 0, 0, 0, 0, loc),
	)

	// CST -> CDT
	check(
		"2018-03-11 01:00:00 -0600 CST",
		NewClock(1, 0),
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
	)
	check(
		"2018-03-11 01:30:00 -0600 CST",
		NewClock(1, 30),
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
	)
	check(
		"2018-03-11 03:00:00 -0500 CDT",
		NewClock(2, 0),
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
	)
	check(
		"2018-03-11 03:00:00 -0500 CDT",
		NewClock(2, 30),
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
	)
	check(
		"2018-03-11 03:00:00 -0500 CDT",
		NewClock(3, 0),
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
	)
	check(
		"2018-03-11 03:30:00 -0500 CDT",
		NewClock(3, 30),
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
	)

	// CDT -> CST
	check(
		"2018-11-04 00:30:00 -0500 CDT",
		NewClock(0, 30),
		time.Date(2018, time.November, 4, 0, 0, 0, 0, loc),
	)
	check(
		"2018-11-04 01:00:00 -0500 CDT",
		NewClock(1, 0),
		time.Date(2018, time.November, 4, 0, 0, 0, 0, loc),
	)
	check(
		"2018-11-04 01:30:00 -0500 CDT",
		NewClock(1, 30),
		time.Date(2018, time.November, 4, 0, 0, 0, 0, loc),
	)
	check(
		"2018-11-04 02:00:00 -0600 CST",
		NewClock(2, 0),
		time.Date(2018, time.November, 4, 0, 0, 0, 0, loc),
	)
	check(
		"2018-11-04 02:30:00 -0600 CST",
		NewClock(2, 30),
		time.Date(2018, time.November, 4, 0, 0, 0, 0, loc),
	)
}

func TestIsDST(t *testing.T) {
	check := func(exp bool, expAt, expChg Clock, ts time.Time) {
		t.Helper()
		is, at, chg := IsDST(ts)
		assert.Equal(t, exp, is)
		assert.Equal(t, expAt.String(), at.String(), "Change Time")
		assert.Equal(t, expChg.String(), chg.String(), "Change Amount")
	}

	loc, err := time.LoadLocation("America/Chicago")
	require.NoError(t, err)
	check(
		false, 0, 0,
		time.Date(2018, time.March, 9, 0, 0, 0, 0, loc),
	)
	check(
		true, NewClock(2, 0), NewClock(1, 0),
		time.Date(2018, time.March, 11, 0, 0, 0, 0, loc),
	)
	check(
		true, NewClock(2, 0), NewClock(-1, 0),
		time.Date(2018, time.November, 4, 0, 0, 0, 0, loc),
	)

}
