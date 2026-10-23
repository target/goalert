package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClock_Days(t *testing.T) {

	check := func(expDays int, expHrs string, c Clock) {
		t.Helper()
		days, c := c.Days()
		assert.Equal(t, expDays, days, "days")
		assert.Equal(t, expHrs, c.String())
	}

	check(2, "12:00", NewClock(60, 0)) // 2 days, 12 hrs
	check(0, "12:00", NewClock(12, 0)) //  12 hrs
	check(1, "00:00", NewClock(24, 0)) // 1 day, 0 hrs

	check(1, "01:00", NewClock(25, 0))  // 1 day, 0 hrs
	check(1, "01:30", NewClock(25, 30)) // 1 day, 0 hrs
}

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

func TestFirstOfDay_MidnightDST(t *testing.T) {
	check := func(expTs string, c Clock, ts time.Time) {
		t.Helper()
		res := c.FirstOfDay(ts)
		assert.Equal(t, expTs, res.String(), "FirstOfDay(%s) on %s", c, ts.Format("2006-01-02"))
	}

	// DST starts at midnight in Santiago on 2026-09-06 (00:00 -> 01:00)
	santiago := mustLoadLocation(t, "America/Santiago")
	day := time.Date(2026, time.September, 6, 12, 0, 0, 0, santiago)
	check("2026-09-06 01:00:00 -0300 -03", NewClock(0, 0), day)
	check("2026-09-06 01:00:00 -0300 -03", NewClock(0, 30), day)
	check("2026-09-06 01:00:00 -0300 -03", NewClock(1, 0), day)
	check("2026-09-06 01:30:00 -0300 -03", NewClock(1, 30), day)
	check("2026-09-06 08:00:00 -0300 -03", NewClock(8, 0), day)
	check("2026-09-06 23:59:00 -0300 -03", NewClock(23, 59), day)
	// the days before and after are unaffected
	check("2026-09-05 08:00:00 -0400 -04", NewClock(8, 0), day.AddDate(0, 0, -1))
	check("2026-09-05 23:59:00 -0400 -04", NewClock(23, 59), day.AddDate(0, 0, -1))
	check("2026-09-07 00:00:00 -0300 -03", NewClock(0, 0), day.AddDate(0, 0, 1))
	check("2026-09-07 08:00:00 -0300 -03", NewClock(8, 0), day.AddDate(0, 0, 1))

	// DST starts at midnight in Havana on 2026-03-08 (00:00 -> 01:00)
	havana := mustLoadLocation(t, "America/Havana")
	day = time.Date(2026, time.March, 8, 12, 0, 0, 0, havana)
	check("2026-03-08 01:00:00 -0400 CDT", NewClock(0, 0), day)
	check("2026-03-08 01:00:00 -0400 CDT", NewClock(0, 30), day)
	check("2026-03-08 08:00:00 -0400 CDT", NewClock(8, 0), day)

	// positive offset: time.Date resolves the missing midnight to 01:00 on the same day
	beirut := mustLoadLocation(t, "Asia/Beirut")
	day = time.Date(2026, time.March, 29, 12, 0, 0, 0, beirut)
	check("2026-03-29 01:00:00 +0300 EEST", NewClock(0, 0), day)
	check("2026-03-29 01:00:00 +0300 EEST", NewClock(0, 30), day)
	check("2026-03-29 08:00:00 +0300 EEST", NewClock(8, 0), day)

	// DST ends at midnight in Santiago on 2026-04-05 (00:00 -> 23:00); 23:xx happens twice on the 4th
	day = time.Date(2026, time.April, 4, 12, 0, 0, 0, santiago)
	check("2026-04-04 22:30:00 -0300 -03", NewClock(22, 30), day)
	check("2026-04-04 23:30:00 -0300 -03", NewClock(23, 30), day)
	check("2026-04-05 00:00:00 -0400 -04", NewClock(0, 0), day.AddDate(0, 0, 1))
	check("2026-04-05 08:00:00 -0400 -04", NewClock(8, 0), day.AddDate(0, 0, 1))

	// DST ends in Havana on 2026-11-01 (01:00 -> 00:00); midnight happens twice
	day = time.Date(2026, time.November, 1, 12, 0, 0, 0, havana)
	check("2026-11-01 00:00:00 -0400 CDT", NewClock(0, 0), day)
	check("2026-11-01 00:30:00 -0400 CDT", NewClock(0, 30), day)
	check("2026-11-01 01:00:00 -0500 CST", NewClock(1, 0), day)
	check("2026-11-01 08:00:00 -0500 CST", NewClock(8, 0), day)
}

func TestLastOfDay_MidnightDST(t *testing.T) {
	check := func(expTs string, c Clock, ts time.Time) {
		t.Helper()
		res := c.LastOfDay(ts)
		assert.Equal(t, expTs, res.String(), "LastOfDay(%s) on %s", c, ts.Format("2006-01-02"))
	}

	// DST starts at midnight in Santiago on 2026-09-06 (00:00 -> 01:00)
	santiago := mustLoadLocation(t, "America/Santiago")
	day := time.Date(2026, time.September, 6, 12, 0, 0, 0, santiago)
	check("2026-09-06 01:00:00 -0300 -03", NewClock(0, 0), day)
	check("2026-09-06 01:00:00 -0300 -03", NewClock(0, 30), day)
	check("2026-09-06 01:00:00 -0300 -03", NewClock(1, 0), day)
	check("2026-09-06 01:30:00 -0300 -03", NewClock(1, 30), day)
	check("2026-09-06 08:00:00 -0300 -03", NewClock(8, 0), day)
	check("2026-09-06 23:59:00 -0300 -03", NewClock(23, 59), day)
	// the days before and after are unaffected
	check("2026-09-05 08:00:00 -0400 -04", NewClock(8, 0), day.AddDate(0, 0, -1))
	check("2026-09-05 23:59:00 -0400 -04", NewClock(23, 59), day.AddDate(0, 0, -1))
	check("2026-09-07 00:00:00 -0300 -03", NewClock(0, 0), day.AddDate(0, 0, 1))
	check("2026-09-07 08:00:00 -0300 -03", NewClock(8, 0), day.AddDate(0, 0, 1))

	// DST starts at midnight in Havana on 2026-03-08 (00:00 -> 01:00)
	havana := mustLoadLocation(t, "America/Havana")
	day = time.Date(2026, time.March, 8, 12, 0, 0, 0, havana)
	check("2026-03-08 01:00:00 -0400 CDT", NewClock(0, 0), day)
	check("2026-03-08 01:00:00 -0400 CDT", NewClock(0, 30), day)
	check("2026-03-08 08:00:00 -0400 CDT", NewClock(8, 0), day)

	// positive offset: time.Date resolves the missing midnight to 01:00 on the same day
	beirut := mustLoadLocation(t, "Asia/Beirut")
	day = time.Date(2026, time.March, 29, 12, 0, 0, 0, beirut)
	check("2026-03-29 01:00:00 +0300 EEST", NewClock(0, 0), day)
	check("2026-03-29 01:00:00 +0300 EEST", NewClock(0, 30), day)
	check("2026-03-29 08:00:00 +0300 EEST", NewClock(8, 0), day)

	// DST ends at midnight in Santiago on 2026-04-05 (00:00 -> 23:00); 23:xx happens twice on the 4th
	day = time.Date(2026, time.April, 4, 12, 0, 0, 0, santiago)
	check("2026-04-04 22:30:00 -0300 -03", NewClock(22, 30), day)
	check("2026-04-04 23:30:00 -0400 -04", NewClock(23, 30), day)
	check("2026-04-05 00:00:00 -0400 -04", NewClock(0, 0), day.AddDate(0, 0, 1))
	check("2026-04-05 08:00:00 -0400 -04", NewClock(8, 0), day.AddDate(0, 0, 1))

	// DST ends in Havana on 2026-11-01 (01:00 -> 00:00); midnight happens twice
	day = time.Date(2026, time.November, 1, 12, 0, 0, 0, havana)
	check("2026-11-01 00:00:00 -0500 CST", NewClock(0, 0), day)
	check("2026-11-01 00:30:00 -0500 CST", NewClock(0, 30), day)
	check("2026-11-01 01:00:00 -0500 CST", NewClock(1, 0), day)
	check("2026-11-01 08:00:00 -0500 CST", NewClock(8, 0), day)
}
