package timeutil

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindNextOffsetChange(t *testing.T) {
	var loc *time.Location
	setLocation := func(newLoc *time.Location, err error) {
		t.Helper()
		loc = newLoc
		require.NoError(t, err)
	}
	check := func(expTime, expClock, start string, dur time.Duration) {
		t.Helper()

		ts, err := time.ParseInLocation("2006-01-02 15:04:05.999999999 -0700 MST", start, loc)
		if err != nil {
			// workaround for being unable to parse "+1030" as "MST"
			parts := strings.Split(start, " ")
			ts, err = time.ParseInLocation("2006-01-02 15:04:05.999999999 -0700", strings.Join(parts[:3], " "), loc)
		}
		require.NoError(t, err)

		at, changeBy := FindNextOffsetChange(ts, dur)

		assert.Equalf(t, expTime, at.String(),
			"time of change within %s of %s (%s)", dur.String(), start, loc.String(),
		)
		assert.Equalf(t, expClock, changeBy.String(),
			"change amount within %s of %s (%s)", dur.String(), start, loc.String(),
		)
	}

	setLocation(time.LoadLocation("America/Chicago"))
	check(
		"2020-03-08 03:00:00 -0500 CDT", "01:00",

		"2020-03-08 00:00:00 -0600 CST", 2*time.Hour,
	)
	check(
		"2020-11-01 01:00:00 -0600 CST", "-1:00",

		"2020-11-01 00:00:00 -0500 CDT", 2*time.Hour,
	)

}

func TestNextClock(t *testing.T) {
	var loc *time.Location
	setLocation := func(newLoc *time.Location, err error) {
		t.Helper()
		loc = newLoc
		require.NoError(t, err)
	}
	check := func(exp, start string, clock Clock) {
		t.Helper()

		ts, err := time.ParseInLocation("2006-01-02 15:04:05.999999999 -0700 MST", start, loc)
		if err != nil {
			// workaround for being unable to parse "+1030" as "MST"
			parts := strings.Split(start, " ")
			ts, err = time.ParseInLocation("2006-01-02 15:04:05.999999999 -0700", strings.Join(parts[:3], " "), loc)
		}
		require.NoError(t, err)

		assert.Equalf(t, exp, NextClock(ts, clock).String(),
			"NextClock of %s from %s (%s)", clock.String(), start, loc.String(),
		)
	}

	setLocation(time.UTC, nil)
	check(
		"2020-01-01 01:00:00 +0000 UTC",
		"2020-01-01 00:00:00 +0000 UTC",
		NewClock(1, 0),
	)
	check(
		"2020-01-02 01:00:00 +0000 UTC",
		"2020-01-01 01:00:00 +0000 UTC",
		NewClock(1, 0),
	)

	setLocation(time.LoadLocation("America/Chicago"))

	// CST -> CDT (spring)
	// Midnight, March 8th 2020 -- at 2:00AM CST the time becomes 3:00AM CDT
	check(
		"2020-03-08 03:00:00 -0500 CDT",
		"2020-03-08 01:00:00 -0600 CST",
		NewClock(2, 0),
	)
	check(
		"2020-03-08 03:00:00 -0500 CDT",
		"2020-03-08 01:00:00 -0600 CST",
		NewClock(2, 30),
	)
	check(
		"2020-03-08 03:30:00 -0500 CDT",
		"2020-03-08 01:00:00 -0600 CST",
		NewClock(3, 30),
	)
	check(
		"2020-03-08 01:45:00 -0600 CST",
		"2020-03-08 01:00:00 -0600 CST",
		NewClock(1, 45),
	)

	// CDT -> CST (fall)
	// Midnight, November 1st 2020 -- at 2:00AM CDT the time becomes 1:00AM CST
	check(
		"2020-11-01 01:30:00 -0500 CDT",
		"2020-11-01 01:00:00 -0500 CDT",
		NewClock(1, 30),
	)
	check(
		"2020-11-01 01:30:00 -0600 CST",
		"2020-11-01 01:30:00 -0500 CDT",
		NewClock(1, 30),
	)
	check(
		"2020-11-01 01:30:00 -0600 CST",
		"2020-11-01 01:35:00 -0500 CDT",
		NewClock(1, 30),
	)
	check(
		"2020-11-01 03:30:00 -0600 CST",
		"2020-11-01 01:35:00 -0500 CDT",
		NewClock(3, 30),
	)
	check(
		"2020-11-01 03:30:00 -0600 CST",
		"2020-11-01 01:35:00 -0600 CST",
		NewClock(3, 30),
	)

}
