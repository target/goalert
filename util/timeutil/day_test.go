package timeutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	require.NoError(t, err)
	return loc
}

func TestStartOfDay(t *testing.T) {
	check := func(exp string, y int, m time.Month, d int, zone string) {
		t.Helper()
		loc := mustLoadLocation(t, zone)
		res := startOfDay(y, m, d, loc)
		assert.Equal(t, exp, res.String(), "startOfDay(%d, %s, %d, %s)", y, m, d, zone)

		// StartOfDay must agree for any time on that day
		assert.Equal(t, exp, StartOfDay(res).String(), "StartOfDay from itself")
		assert.Equal(t, exp, StartOfDay(time.Date(y, m, d, 12, 0, 0, 0, loc)).String(), "StartOfDay from noon")
		assert.Equal(t, exp, StartOfDay(time.Date(y, m, d, 23, 59, 0, 0, loc)).String(), "StartOfDay from 23:59")
	}

	// normal days
	check("2026-09-04 00:00:00 -0400 -04", 2026, time.September, 4, "America/Santiago")
	check("2026-07-04 00:00:00 -0500 CDT", 2026, time.July, 4, "America/Chicago")
	check("2026-07-04 00:00:00 +0000 UTC", 2026, time.July, 4, "UTC")

	// DST starts at midnight (00:00 -> 01:00), so local midnight does not exist.
	// time.Date resolves it to 23:00 the previous day in these zones.
	check("2026-09-06 01:00:00 -0300 -03", 2026, time.September, 6, "America/Santiago")
	check("2026-03-08 01:00:00 -0400 CDT", 2026, time.March, 8, "America/Havana")
	check("2026-03-29 01:00:00 +0000 +00", 2026, time.March, 29, "Atlantic/Azores")
	// positive offset: time.Date already resolves to 01:00 on the same day
	check("2026-03-29 01:00:00 +0300 EEST", 2026, time.March, 29, "Asia/Beirut")

	// DST ends at midnight (00:00 -> 23:00 the previous day); the 4th is 25 hours long
	check("2026-04-04 00:00:00 -0300 -03", 2026, time.April, 4, "America/Santiago")
	check("2026-04-05 00:00:00 -0400 -04", 2026, time.April, 5, "America/Santiago")

	// DST ends 01:00 -> 00:00, so midnight exists twice; the first is returned
	check("2026-11-01 00:00:00 -0400 CDT", 2026, time.November, 1, "America/Havana")
	check("2026-10-25 00:00:00 +0000 +00", 2026, time.October, 25, "Atlantic/Azores")

	// day values are normalized like time.Date
	check("2026-09-06 01:00:00 -0300 -03", 2026, time.August, 37, "America/Santiago")
	check("2026-09-06 01:00:00 -0300 -03", 2026, time.October, -24, "America/Santiago")
	check("2026-02-28 00:00:00 -0600 CST", 2026, time.March, 0, "America/Chicago")
	check("2027-01-01 00:00:00 -0600 CST", 2026, time.December, 32, "America/Chicago")
}

func TestStartOfDay_MatchesDate(t *testing.T) {
	// For every day of 2026 in zones where local midnight always exists,
	// startOfDay must be identical to time.Date.
	for _, zone := range []string{"America/Chicago", "Europe/London", "Pacific/Auckland", "Australia/Lord_Howe", "UTC"} {
		loc := mustLoadLocation(t, zone)
		for day := 1; day <= 365; day++ {
			exp := time.Date(2026, time.January, day, 0, 0, 0, 0, loc)
			require.Equal(t, day, exp.YearDay(), "%s: midnight should exist on day %d", zone, day)

			y, m, d := exp.Date()
			res := startOfDay(y, m, d, loc)
			require.Truef(t, exp.Equal(res), "%s %s: got %s; want %s", zone, exp.Format("2006-01-02"), res, exp)
			require.Equal(t, loc, res.Location(), zone)
		}
	}

	// In zones with a DST transition at midnight, startOfDay is always on the
	// requested day, and only differs from midnight on the day of the gap.
	gapDays := map[string]string{
		"America/Santiago": "2026-09-06",
		"America/Havana":   "2026-03-08",
		"Atlantic/Azores":  "2026-03-29",
		"Asia/Beirut":      "2026-03-29",
	}
	for zone, gapDay := range gapDays {
		loc := mustLoadLocation(t, zone)
		for day := 1; day <= 365; day++ {
			y, m, d := time.Date(2026, time.January, day, 12, 0, 0, 0, loc).Date()
			res := startOfDay(y, m, d, loc)
			date := fmt.Sprintf("%04d-%02d-%02d", y, m, d)
			require.Equal(t, date, res.Format("2006-01-02"), "%s: startOfDay should be on the requested day", zone)
			if date == gapDay {
				assert.Equal(t, "01:00:00", res.Format("15:04:05"), "%s %s: should start after the gap", zone, date)
				continue
			}

			exp := time.Date(y, m, d, 0, 0, 0, 0, loc)
			require.Truef(t, exp.Equal(res), "%s %s: got %s; want %s", zone, date, res, exp)
		}
	}
}

func TestAddDays(t *testing.T) {
	check := func(exp string, ts time.Time, days int) {
		t.Helper()
		res := AddDays(ts, days)
		assert.Equal(t, exp, res.String(), "AddDays(%s, %d)", ts, days)
	}

	chicago := mustLoadLocation(t, "America/Chicago")
	santiago := mustLoadLocation(t, "America/Santiago")

	// normal; wall-clock value (including sub-second) is preserved
	check("2026-07-05 08:30:15.5 -0500 CDT", time.Date(2026, time.July, 4, 8, 30, 15, 500_000_000, chicago), 1)
	check("2026-07-03 08:30:15.5 -0500 CDT", time.Date(2026, time.July, 4, 8, 30, 15, 500_000_000, chicago), -1)
	check("2026-08-03 08:30:00 -0500 CDT", time.Date(2026, time.July, 4, 8, 30, 0, 0, chicago), 30)
	check("2026-07-04 08:30:00 -0500 CDT", time.Date(2026, time.July, 4, 8, 30, 0, 0, chicago), 0)

	// across a DST change (02:00 -> 03:00 on 2026-03-08), the wall-clock value is preserved
	check("2026-03-08 08:30:00 -0500 CDT", time.Date(2026, time.March, 7, 8, 30, 0, 0, chicago), 1)
	// 02:30 does not exist on 2026-03-08, so the first instant after is returned
	check("2026-03-08 03:00:00 -0500 CDT", time.Date(2026, time.March, 7, 2, 30, 0, 0, chicago), 1)
	// 01:30 exists twice on 2026-11-01, so the first is returned
	check("2026-11-01 01:30:00 -0500 CDT", time.Date(2026, time.October, 31, 1, 30, 0, 0, chicago), 1)

	// DST starts at midnight in Santiago on 2026-09-06 (00:00 -> 01:00)
	check("2026-09-06 08:00:00 -0300 -03", time.Date(2026, time.September, 5, 8, 0, 0, 0, santiago), 1)
	check("2026-09-06 08:00:00 -0300 -03", time.Date(2026, time.September, 7, 8, 0, 0, 0, santiago), -1)
	// 00:30 does not exist on the 6th, so the first instant of the day is returned
	check("2026-09-06 01:00:00 -0300 -03", time.Date(2026, time.September, 5, 0, 30, 0, 0, santiago), 1)
	check("2026-09-06 01:00:00 -0300 -03", time.Date(2026, time.September, 7, 0, 30, 0, 0, santiago), -1)
	// from the first instant of the gap day
	check("2026-09-05 01:00:00 -0400 -04", time.Date(2026, time.September, 6, 1, 0, 0, 0, santiago), -1)
	check("2026-09-07 01:00:00 -0300 -03", time.Date(2026, time.September, 6, 1, 0, 0, 0, santiago), 1)
}
