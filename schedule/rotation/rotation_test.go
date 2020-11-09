package rotation

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const timeFmt = "Jan 2 2006 3:04 pm"

func mustParse(t *testing.T, value string) time.Time {
	t.Helper()
	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		t.Fatal(err)
	}
	tm, err := time.ParseInLocation(timeFmt, value, loc)
	if err != nil {
		t.Fatal(err)
	}
	tm = tm.In(loc)
	return tm
}

func TestAddClockHours(t *testing.T) {
	var loc *time.Location
	setLocation := func(newLoc *time.Location, err error) {
		t.Helper()
		loc = newLoc
		require.NoError(t, err)
	}
	check := func(exp, start string, hours int) {
		t.Helper()

		ts, err := time.ParseInLocation("2006-01-02 15:04:05.999999999 -0700 MST", start, loc)
		if err != nil {
			// workaround for being unable to parse "+1030" as "MST"
			parts := strings.Split(start, " ")
			ts, err = time.ParseInLocation("2006-01-02 15:04:05.999999999 -0700", strings.Join(parts[:3], " "), loc)
		}
		require.NoError(t, err)

		op := "add"
		op2 := "to"
		hoursV := hours
		if hours < 0 {
			op = "subtract"
			op2 = "from"
			hoursV = -hours
		}
		p := "s"
		if hoursV == 1 {
			p = ""
		}

		assert.Equalf(t, exp, addClockHours(ts, hours).String(),
			"%s %d hour%s %s %s (%s)", op, hours, p, op2, start, loc.String(),
		)
	}

	// UTC basics
	setLocation(time.UTC, nil)
	check(
		"2020-01-01 01:00:00 +0000 UTC",
		"2020-01-01 00:00:00 +0000 UTC", 1,
	)
	check(
		"2020-01-01 04:00:00 +0000 UTC",
		"2020-01-01 00:00:00 +0000 UTC", 4,
	)
	check(
		"2020-01-01 00:00:00 +0000 UTC",
		"2020-01-01 01:00:00 +0000 UTC", -1,
	)
	check(
		"2020-01-01 00:00:00 +0000 UTC",
		"2020-01-01 04:00:00 +0000 UTC", -4,
	)

	// Leap second @ 23:59:60
	check(
		"2017-01-01 00:00:00 +0000 UTC",
		"2016-12-31 23:00:00 +0000 UTC", 1,
	)
	check(
		"2016-12-31 23:00:00 +0000 UTC",
		"2017-01-01 00:00:00 +0000 UTC", -1,
	)

	// CST -> CDT (spring)
	// // Midnight, March 8th 2020 -- at 2:00AM CST the time becomes 3:00AM CDT
	setLocation(time.LoadLocation("America/Chicago"))

	// Adding 1 clock hour to 1:00AM CST should result in 3:00AM CDT due to the
	// transition at 2:00AM CST (2:00AM doesn't exist outside of ambiguity)
	check(
		"2020-03-08 03:00:00 -0500 CDT",
		"2020-03-08 01:00:00 -0600 CST", 1,
	)

	// Adding 2 clock hours to 1:00AM CST should also result in 3:00AM CDT
	check(
		"2020-03-08 03:00:00 -0500 CDT",
		"2020-03-08 01:00:00 -0600 CST", 2,
	)

	// Adding 4 clock hours to 12:00AM CST should result in 4:00AM CDT
	check(
		"2020-03-08 04:00:00 -0500 CDT",
		"2020-03-08 00:00:00 -0600 CST", 4,
	)

	// Two hours past 12:00AM CST would be 3:00AM CDT
	// subtracting 1 hour would be 1:00AM CST, reversing the transition
	check(
		"2020-03-08 01:00:00 -0600 CST",
		"2020-03-08 03:00:00 -0500 CDT", -1,
	)

	// Likewise, subtracting 2 clock-hours would also be 1:00AM CST
	check(
		"2020-03-08 01:00:00 -0600 CST",
		"2020-03-08 03:00:00 -0500 CDT", -2,
	)

	// Lastly, subtracting 4 clock-hours from 4:00AM CDT would be 12:00AM CST.
	// 4:00AM CDT would be 3 real-time hours past midnight.
	check(
		"2020-03-08 00:00:00 -0600 CST",
		"2020-03-08 04:00:00 -0500 CDT", -4,
	)

	// CDT -> CST (fall)
	// Midnight, November 1st 2020 -- at 2:00AM CDT the time becomes 1:00AM CST
	setLocation(time.LoadLocation("America/Chicago"))

	// Adding 1 clock hour to 1:00AM CDT should result in 2:00AM CST due to the
	// transition at 2:00AM (1:00AM repeats, first in CDT, then in CST)
	// so to advance the clock 2 hours end up passing.
	check(
		"2020-11-01 02:00:00 -0600 CST",
		"2020-11-01 01:00:00 -0500 CDT", 1,
	)

	// Adding 2 clock hours to 1:00AM CDT should result in 3:00AM CDT
	check(
		"2020-11-01 03:00:00 -0600 CST",
		"2020-11-01 01:00:00 -0500 CDT", 2,
	)

	// Adding 4 clock hours to 12:00AM CDT should result in 4:00AM CST
	check(
		"2020-11-01 04:00:00 -0600 CST",
		"2020-11-01 00:00:00 -0500 CDT", 4,
	)

	// Two hours past 12:00AM CDT would be 1:00AM CST
	// subtracting 1 clock-hour would be 12:00AM CDT, reversing the transition
	check(
		"2020-11-01 00:00:00 -0500 CDT",
		"2020-11-01 01:00:00 -0600 CST", -1,
	)

	// subtracting 1 clock-hour from 1:00AM CDT would also be 12:00AM CDT
	check(
		"2020-11-01 00:00:00 -0500 CDT",
		"2020-11-01 01:00:00 -0500 CDT", -1,
	)

	// Subtracting 2 clock-hours from 1:00AM CDT would be Oct 31st 11:00PM CDT
	check(
		"2020-10-31 23:00:00 -0500 CDT",
		"2020-11-01 01:00:00 -0500 CDT", -2,
	)

	// Subtracting 2 clock-hours from 1:00AM CST would also be Oct 31st 11:00PM CDT
	check(
		"2020-10-31 23:00:00 -0500 CDT",
		"2020-11-01 01:00:00 -0600 CST", -2,
	)

	// Lastly, subtracting 4 clock-hours from 4:00AM CST should be 12:00AM CDT.
	// 4:00AM CST would be 5 real-time hours past midnight.
	check(
		"2020-11-01 00:00:00 -0500 CDT",
		"2020-11-01 04:00:00 -0600 CST", -4,
	)

	// fancy edge case (30 minute DST)
	// 12:00AM April 5th 2020 -- at 2:00AM time becomes 1:30AM
	setLocation(time.LoadLocation("Australia/Lord_Howe"))

	// Adding 1 clock hour to 1:00AM should result in 2:00AM. Due to the
	// transition at 2:00AM (1:30AM repeats), 1.5 hours end up passing.
	check(
		"2020-04-05 02:00:00 +1030 +1030",
		"2020-04-05 01:00:00 +1100 +11", 1,
	)

	// Adding 4 clock-hours should result in 4:00AM +1030 and 4.5 hours passing.
	check(
		"2020-04-05 04:00:00 +1030 +1030",
		"2020-04-05 00:00:00 +1100 +11", 4,
	)

	// Subtracting 4 clock-hours from 4:00AM +1030 should result in 12:00AM +11
	check(
		"2020-04-05 00:00:00 +1100 +11",
		"2020-04-05 04:00:00 +1030 +1030", -4,
	)
	// 12:00AM Oct 4th 2020 -- at 2:00AM time becomes 2:30AM
	setLocation(time.LoadLocation("Australia/Lord_Howe"))

	// Adding 1 clock hour to 1:00AM should result in 2:30AM. Due to the
	// transition at 2:00AM.
	check(
		"2020-10-04 02:30:00 +1100 +11",
		"2020-10-04 01:00:00 +1030 +1030", 1,
	)

	// Adding 4 clock-hours should result in 4:00AM +1030 and 4.5 hours passing.
	check(
		"2020-04-05 04:00:00 +1030 +1030",
		"2020-04-05 00:00:00 +1100 +11", 4,
	)

	// Subtracting 4 clock-hours from 4:00AM +1030 should result in 12:00AM +11
	check(
		"2020-04-05 00:00:00 +1100 +11",
		"2020-04-05 04:00:00 +1030 +1030", -4,
	)

}

func TestRotation_EndTime_DST(t *testing.T) {
	tFmt := timeFmt + " (-07:00)"
	rot := &Rotation{
		Type:  TypeHourly,
		Start: mustParse(t, "Jan 1 2017 1:00 am"),
	}
	t.Logf("Rotation Start=%s", rot.Start.Format(tFmt))

	test := func(start, end time.Time) {
		t.Helper()
		t.Run("", func(t *testing.T) {
			t.Helper()
			t.Logf("Shift Start=%s", start.Format(tFmt))
			e := rot.EndTime(start)
			if !e.Equal(end) {
				t.Errorf("got '%s' want '%s'", e.Format(tFmt), end.Format(tFmt))
			}
		})
	}

	start := rot.Start.AddDate(0, 2, 11) // mar 11 1:00am
	end := start.Add(time.Hour)          // same time (we skip a shift)
	test(start, end)

	start = rot.Start.AddDate(0, 10, 4) // nov 5 1:00am
	end = start.Add(time.Hour * 2)      // 2 hours after
	test(start, end)
}

func TestRotation_EndTime_ConfigChange(t *testing.T) {
	rot := &Rotation{
		Type:        TypeHourly,
		Start:       mustParse(t, "Jan 1 2017 12:00 am"),
		ShiftLength: 12,
	}

	start := mustParse(t, "Jan 3 2017 6:00 am")
	result := rot.EndTime(start)

	expected := mustParse(t, "Jan 3 2017 12:00 pm")
	if !result.Equal(expected) {
		t.Errorf("EndTime=%s; want %s", result.Format(timeFmt), expected.Format(timeFmt))
	}
}

func TestRotation_EndTime(t *testing.T) {
	test := func(start, end string, len int, dur time.Duration, typ Type) {
		t.Run(string(typ), func(t *testing.T) {
			s := mustParse(t, start)
			e := mustParse(t, end)
			dur = dur.Round(time.Second)
			if e.Sub(s).Round(time.Second) != dur {
				t.Fatalf("bad test data: end-start=%s; want %s", e.Sub(s).Round(time.Second).String(), dur.String())
			}
			rot := &Rotation{
				Type:        typ,
				ShiftLength: len,
				Start:       s,
			}

			result := rot.EndTime(s)
			if !result.Equal(e) {
				t.Errorf("got '%s'; want '%s'", result.Format(timeFmt), end)
			}
			if result.Sub(s).Round(time.Second) != dur {
				t.Errorf("duration was '%s'; want '%s'", result.Sub(s).Round(time.Second).String(), dur.String())
			}
		})
	}

	type dat struct {
		s   string
		l   int
		exp string
		dur time.Duration
	}

	// weekly
	data := []dat{
		{s: "Jun 10 2017 8:00 am", l: 1, exp: "Jun 17 2017 8:00 am", dur: time.Hour * 24 * 7},
		{s: "Jun 10 2017 8:00 am", l: 2, exp: "Jun 24 2017 8:00 am", dur: time.Hour * 24 * 7 * 2},

		// DST tests
		{s: "Mar 10 2017 8:00 am", l: 1, exp: "Mar 17 2017 8:00 am", dur: time.Hour*24*7 - time.Hour},
		{s: "Nov 4 2017 8:00 am", l: 1, exp: "Nov 11 2017 8:00 am", dur: time.Hour*24*7 + time.Hour},
	}
	for _, d := range data {
		test(d.s, d.exp, d.l, d.dur, TypeWeekly)
	}

	// daily
	data = []dat{
		{s: "Jun 10 2017 8:00 am", l: 1, exp: "Jun 11 2017 8:00 am", dur: time.Hour * 24},
		{s: "Jun 10 2017 8:00 am", l: 2, exp: "Jun 12 2017 8:00 am", dur: time.Hour * 24 * 2},

		// DST tests
		{s: "Mar 11 2017 8:00 am", l: 1, exp: "Mar 12 2017 8:00 am", dur: time.Hour * 23},
		{s: "Nov 4 2017 8:00 am", l: 1, exp: "Nov 5 2017 8:00 am", dur: time.Hour * 25},
	}
	for _, d := range data {
		test(d.s, d.exp, d.l, d.dur, TypeDaily)
	}

	// hourly
	data = []dat{
		{s: "Jun 10 2017 8:00 am", l: 1, exp: "Jun 10 2017 9:00 am", dur: time.Hour},
		{s: "Jun 10 2017 8:00 am", l: 2, exp: "Jun 10 2017 10:00 am", dur: time.Hour * 2},

		// DST tests
		{s: "Mar 12 2017 12:00 am", l: 3, exp: "Mar 12 2017 3:00 am", dur: time.Hour * 2},
		{s: "Nov 5 2017 12:00 am", l: 3, exp: "Nov 5 2017 3:00 am", dur: time.Hour * 4},
	}
	for _, d := range data {
		test(d.s, d.exp, d.l, d.dur, TypeHourly)
	}
}

func TestRotation_Normalize(t *testing.T) {

	test := func(valid bool, r Rotation) {
		name := "valid"
		if !valid {
			name = "invalid"
		}
		t.Run(name, func(t *testing.T) {
			_, err := r.Normalize()
			if valid && err != nil {
				t.Errorf("err = %v; want nil", err)
			} else if !valid && err == nil {
				t.Errorf("got nil err; want non-nil")
			}
		})
	}

	valid := []Rotation{
		{Name: "Default", ShiftLength: 1, Type: TypeWeekly, Description: "Default Rotation"},
	}
	invalid := []Rotation{
		{Name: "D", ShiftLength: -100, Type: TypeWeekly, Description: "Default Rotation"},
	}
	for _, r := range valid {
		test(true, r)
	}
	for _, r := range invalid {
		test(false, r)
	}
}

func TestRotation_FutureStart(t *testing.T) {
	rot := Rotation{
		Type:        TypeDaily,
		ShiftLength: 1,

		// StartTime and EndTime should work correctly even if Start
		// is in the future.
		Start: time.Date(2019, 0, 10, 0, 0, 0, 0, time.UTC),
	}

	assert.Equal(t, time.Date(2019, 0, 6, 0, 0, 0, 0, time.UTC),
		rot.EndTime(time.Date(2019, 0, 5, 0, 0, 0, 0, time.UTC)),
	)
	assert.Equal(t, time.Date(2019, 0, 5, 0, 0, 0, 0, time.UTC),
		rot.StartTime(time.Date(2019, 0, 5, 0, 0, 0, 0, time.UTC)),
	)
	assert.Equal(t, time.Date(2019, 0, 4, 0, 0, 0, 0, time.UTC),
		rot.StartTime(time.Date(2019, 0, 5, 0, 0, 0, -1, time.UTC)),
	)
	assert.Equal(t, time.Date(2019, 0, 11, 0, 0, 0, 0, time.UTC),
		rot.EndTime(time.Date(2019, 0, 10, 0, 0, 0, 0, time.UTC)),
	)
	assert.Equal(t, time.Date(2019, 0, 10, 0, 0, 0, 0, time.UTC),
		rot.StartTime(time.Date(2019, 0, 10, 0, 0, 0, 0, time.UTC)),
	)
	assert.Equal(t, time.Date(2019, 0, 9, 0, 0, 0, 0, time.UTC),
		rot.StartTime(time.Date(2019, 0, 10, 0, 0, 0, -1, time.UTC)),
	)
}

func TestRotation_StartTime(t *testing.T) {
	test := func(start, end string, len int, dur time.Duration, typ Type) {
		t.Run(string(typ), func(t *testing.T) {
			s := mustParse(t, start)
			e := mustParse(t, end)
			dur = dur.Round(time.Second)
			if e.Sub(s).Round(time.Second) != dur {
				t.Fatalf("bad test data: end-start=%s; want %s", e.Sub(s).Round(time.Second).String(), dur.String())
			}
			rot := &Rotation{
				Type:        typ,
				ShiftLength: len,
				Start:       s,
			}

			result := rot.StartTime(s)
			if !result.Equal(s) {
				t.Errorf("got '%s'; want '%s'", result.Format(timeFmt), start)
			}

			if result.Sub(e).Round(time.Second) != -(dur) {
				t.Errorf("duration was '%s'; want '%s'", result.Sub(e).Round(time.Second).String(), dur.String())
			}
		})
	}

	type dat struct {
		s   string
		l   int
		exp string
		dur time.Duration
	}

	// weekly
	data := []dat{
		{s: "Jun 10 2017 8:00 am", l: 1, exp: "Jun 17 2017 8:00 am", dur: time.Hour * 24 * 7},
		{s: "Jun 10 2017 8:00 am", l: 2, exp: "Jun 24 2017 8:00 am", dur: time.Hour * 24 * 7 * 2},

		// DST tests
		{s: "Mar 10 2017 8:00 am", l: 1, exp: "Mar 17 2017 8:00 am", dur: time.Hour*24*7 - time.Hour},
		{s: "Nov 4 2017 8:00 am", l: 1, exp: "Nov 11 2017 8:00 am", dur: time.Hour*24*7 + time.Hour},
	}
	for _, d := range data {
		test(d.s, d.exp, d.l, d.dur, TypeWeekly)
	}

	// weekly but with different start timestamp for comparison
	data = []dat{
		{s: "Jun 10 2017 8:00 am", l: 1, exp: "Jun 17 2017 8:00 am", dur: time.Hour * 24 * 7},
	}
	for _, d := range data {
		test("Jun 16 2017 8:00 am", d.exp, d.l, time.Hour*24, TypeWeekly)
	}

	// daily
	data = []dat{
		{s: "Jun 10 2017 8:00 am", l: 1, exp: "Jun 11 2017 8:00 am", dur: time.Hour * 24},
		{s: "Jun 10 2017 8:00 am", l: 2, exp: "Jun 12 2017 8:00 am", dur: time.Hour * 24 * 2},

		// DST tests
		{s: "Mar 11 2017 8:00 am", l: 1, exp: "Mar 12 2017 8:00 am", dur: time.Hour * 23},
		{s: "Nov 4 2017 8:00 am", l: 1, exp: "Nov 5 2017 8:00 am", dur: time.Hour * 25},
	}
	for _, d := range data {
		test(d.s, d.exp, d.l, d.dur, TypeDaily)
	}

	// hourly
	data = []dat{
		{s: "Jun 10 2017 8:00 am", l: 1, exp: "Jun 10 2017 9:00 am", dur: time.Hour},
		{s: "Jun 10 2017 8:00 am", l: 2, exp: "Jun 10 2017 10:00 am", dur: time.Hour * 2},

		// DST tests
		{s: "Mar 12 2017 12:00 am", l: 3, exp: "Mar 12 2017 3:00 am", dur: time.Hour * 2},
		{s: "Nov 5 2017 12:00 am", l: 3, exp: "Nov 5 2017 3:00 am", dur: time.Hour * 4},
	}
	for _, d := range data {
		test(d.s, d.exp, d.l, d.dur, TypeHourly)
	}
}
