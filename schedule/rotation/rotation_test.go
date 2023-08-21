package rotation

import (
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

func TestRotation_StartEnd_BruteForce(t *testing.T) {
	loc, err := time.LoadLocation("America/Chicago")
	require.NoError(t, err)

	// check will walk through, in 30 sec. increments, from the start to the end time calling
	// both StartTime and EndTime on the rotation and asserting it gives the exact expectedHandoffs.
	check := func(rot *Rotation, start, end time.Time, expectedHandoffs ...string) {
		t.Helper()
		ts := start
		require.Equalf(t, expectedHandoffs[1], rot.EndTime(ts).String(), "first shift end %s", ts.String())
		require.Equalf(t, expectedHandoffs[0], rot.StartTime(ts).String(), "first shift start %s", ts.String())
		timesM := make(map[string]bool)
		var times []string
		for !ts.After(end) {
			res := rot.StartTime(ts).String()
			if !timesM[res] {
				timesM[res] = true
				times = append(times, res)
			}
			require.Contains(t, expectedHandoffs, res, "StartTime for %s", ts.String())

			res = rot.EndTime(ts).String()
			if !timesM[res] {
				timesM[res] = true
				times = append(times, res)
			}
			require.Contains(t, expectedHandoffs, res, "EndTime for %s", ts.String())

			ts = ts.Add(30 * time.Second)
		}
		assert.EqualValues(t, expectedHandoffs, times)
	}

	check(&Rotation{
		Type:        TypeHourly,
		ShiftLength: 24,
		Start:       time.Date(2020, time.October, 30, 1, 30, 0, 0, loc),
	},
		time.Date(2020, time.October, 30, 1, 0, 0, 0, loc),
		time.Date(2020, time.November, 3, 1, 0, 0, 0, loc),

		"2020-10-29 01:30:00 -0500 CDT", // StartTime as of 1AM on the 30th would be 1:30AM on the 29th
		"2020-10-30 01:30:00 -0500 CDT",
		"2020-10-31 01:30:00 -0500 CDT",
		"2020-11-01 01:30:00 -0500 CDT",
		"2020-11-02 01:30:00 -0600 CST",
		"2020-11-03 01:30:00 -0600 CST",
	)

	// Daily & 24 hour should be identical
	check(&Rotation{
		Type:        TypeDaily,
		ShiftLength: 1,
		Start:       time.Date(2020, time.October, 30, 1, 30, 0, 0, loc),
	},
		time.Date(2020, time.October, 30, 1, 0, 0, 0, loc),
		time.Date(2020, time.November, 3, 1, 0, 0, 0, loc),

		"2020-10-29 01:30:00 -0500 CDT",
		"2020-10-30 01:30:00 -0500 CDT",
		"2020-10-31 01:30:00 -0500 CDT",
		"2020-11-01 01:30:00 -0500 CDT",
		"2020-11-02 01:30:00 -0600 CST",
		"2020-11-03 01:30:00 -0600 CST",
	)

	check(&Rotation{
		Type:        TypeDaily,
		ShiftLength: 1,
		Start:       time.Date(2020, time.March, 6, 2, 30, 0, 0, loc),
	},
		time.Date(2020, time.March, 6, 1, 0, 0, 0, loc),
		time.Date(2020, time.March, 10, 1, 0, 0, 0, loc),

		"2020-03-05 02:30:00 -0600 CST", // StartTime as of 1AM on the 6th would be 1:30AM on the 5th
		"2020-03-06 02:30:00 -0600 CST",
		"2020-03-07 02:30:00 -0600 CST",
		"2020-03-08 03:00:00 -0500 CDT",
		"2020-03-09 02:30:00 -0500 CDT",
		"2020-03-10 02:30:00 -0500 CDT",
	)

	check(&Rotation{
		Type:        TypeHourly,
		ShiftLength: 1,
		Start:       time.Date(2020, time.March, 6, 2, 30, 0, 0, loc),
	},
		time.Date(2020, time.March, 8, 1, 0, 0, 0, loc),
		time.Date(2020, time.March, 8, 4, 0, 0, 0, loc),

		"2020-03-08 00:30:00 -0600 CST",
		"2020-03-08 01:30:00 -0600 CST",
		"2020-03-08 03:00:00 -0500 CDT",
		"2020-03-08 03:30:00 -0500 CDT",
		"2020-03-08 04:30:00 -0500 CDT",
	)
	check(&Rotation{
		Type:        TypeHourly,
		ShiftLength: 1,
		Start:       time.Date(2020, time.March, 6, 2, 0, 0, 0, loc),
	},
		time.Date(2020, time.March, 8, 1, 0, 0, 0, loc),
		time.Date(2020, time.March, 8, 4, 0, 0, 0, loc),

		"2020-03-08 01:00:00 -0600 CST",
		"2020-03-08 03:00:00 -0500 CDT",
		"2020-03-08 04:00:00 -0500 CDT",
		"2020-03-08 05:00:00 -0500 CDT",
	)

	check(&Rotation{
		Type:        TypeHourly,
		ShiftLength: 1,
		Start:       time.Date(2020, time.July, 6, 2, 30, 0, 0, loc),
	},
		time.Date(2020, time.November, 1, 0, 0, 0, 0, loc),
		time.Date(2020, time.November, 1, 4, 0, 0, 0, loc),

		"2020-10-31 23:30:00 -0500 CDT",
		"2020-11-01 00:30:00 -0500 CDT",
		"2020-11-01 01:30:00 -0500 CDT",
		"2020-11-01 02:30:00 -0600 CST",
		"2020-11-01 03:30:00 -0600 CST",
		"2020-11-01 04:30:00 -0600 CST",
	)
	check(&Rotation{
		Type:        TypeHourly,
		ShiftLength: 1,
		Start:       time.Date(2020, time.July, 6, 2, 0, 0, 0, loc),
	},
		time.Date(2020, time.November, 1, 0, 0, 0, 0, loc),
		time.Date(2020, time.November, 1, 4, 0, 0, 0, loc),

		"2020-11-01 00:00:00 -0500 CDT",
		"2020-11-01 01:00:00 -0500 CDT",
		"2020-11-01 02:00:00 -0600 CST",
		"2020-11-01 03:00:00 -0600 CST",
		"2020-11-01 04:00:00 -0600 CST",
		"2020-11-01 05:00:00 -0600 CST",
	)

	// monthly rotation sanity check
	check(&Rotation{
		Type:        TypeMonthly,
		ShiftLength: 1,
		Start:       time.Date(2020, time.January, 2, 1, 30, 0, 0, loc),
	},
		time.Date(2020, time.April, 5, 1, 0, 0, 0, loc),
		time.Date(2020, time.June, 12, 1, 0, 0, 0, loc),

		"2020-04-02 01:30:00 -0500 CDT",
		"2020-05-02 01:30:00 -0500 CDT",
		"2020-06-02 01:30:00 -0500 CDT",
		"2020-07-02 01:30:00 -0500 CDT",
	)

	// check monthly rotations with shift from daylight savings to standard time
	check(&Rotation{
		Type:        TypeMonthly,
		ShiftLength: 1,
		Start:       time.Date(2020, time.October, 1, 1, 30, 0, 0, loc),
	},
		time.Date(2020, time.October, 2, 1, 0, 0, 0, loc),
		time.Date(2020, time.December, 2, 1, 0, 0, 0, loc),

		"2020-10-01 01:30:00 -0500 CDT",
		"2020-11-01 01:30:00 -0500 CDT",
		"2020-12-01 01:30:00 -0600 CST",
		"2021-01-01 01:30:00 -0600 CST",
	)

	loc, err = time.LoadLocation("Australia/Lord_Howe")
	require.NoError(t, err)

	// check when DT ends using a non-cst tz (hourly)
	check(&Rotation{
		Type:        TypeHourly,
		ShiftLength: 1,
		Start:       time.Date(2021, time.April, 3, 2, 30, 0, 0, loc),
	},
		time.Date(2021, time.April, 4, 1, 0, 0, 0, loc),
		time.Date(2021, time.April, 4, 4, 0, 0, 0, loc),

		"2021-04-04 00:30:00 +1100 +11",
		"2021-04-04 01:30:00 +1100 +11",   // clock goes back, extra 30 minutes on-call from 1:30-2
		"2021-04-04 02:30:00 +1030 +1030", // new shift starts at 2:30
		"2021-04-04 03:30:00 +1030 +1030",
		"2021-04-04 04:30:00 +1030 +1030",
	)

	// DST ending using a tz with a 30 minute offset change (daily)
	// handoff at end
	check(&Rotation{
		Type:        TypeDaily,
		ShiftLength: 1,
		Start:       time.Date(2021, time.April, 1, 2, 30, 0, 0, loc),
	},
		time.Date(2021, time.April, 2, 1, 0, 0, 0, loc),
		time.Date(2021, time.April, 6, 1, 0, 0, 0, loc),

		"2021-04-01 02:30:00 +1100 +11",
		"2021-04-02 02:30:00 +1100 +11",
		"2021-04-03 02:30:00 +1100 +11",
		"2021-04-04 02:30:00 +1030 +1030", // new offset changes 30 minutes
		"2021-04-05 02:30:00 +1030 +1030",
		"2021-04-06 02:30:00 +1030 +1030",
	)

	// DST ending using a tz with a 30 minute offset change
	// handoff at start
	check(&Rotation{
		Type:        TypeDaily,
		ShiftLength: 1,
		Start:       time.Date(2021, time.April, 2, 2, 0, 0, 0, loc),
	},
		time.Date(2021, time.April, 2, 1, 0, 0, 0, loc),
		time.Date(2021, time.April, 6, 1, 0, 0, 0, loc),

		"2021-04-01 02:00:00 +1100 +11",
		"2021-04-02 02:00:00 +1100 +11",
		"2021-04-03 02:00:00 +1100 +11",
		"2021-04-04 02:00:00 +1030 +1030", // TZ switches at start, +30 minutes on-call for this shift
		"2021-04-05 02:00:00 +1030 +1030",
		"2021-04-06 02:00:00 +1030 +1030",
	)

	// daily as 24 hours - sanity check
	check(&Rotation{
		Type:        TypeHourly,
		ShiftLength: 24,
		Start:       time.Date(2021, time.April, 2, 2, 0, 0, 0, loc),
	},
		time.Date(2021, time.April, 2, 1, 0, 0, 0, loc),
		time.Date(2021, time.April, 6, 1, 0, 0, 0, loc),

		"2021-04-01 02:00:00 +1100 +11",
		"2021-04-02 02:00:00 +1100 +11",
		"2021-04-03 02:00:00 +1100 +11",
		"2021-04-04 02:00:00 +1030 +1030", // TZ switches at start, +30 minutes on-call for this shift
		"2021-04-05 02:00:00 +1030 +1030",
		"2021-04-06 02:00:00 +1030 +1030",
	)

	// check when DST starts for non-CST timezone (Lord_Howe)
	check(&Rotation{
		Type:        TypeHourly,
		ShiftLength: 24,
		Start:       time.Date(2020, time.October, 2, 1, 30, 0, 0, loc),
	},
		time.Date(2020, time.October, 2, 1, 0, 0, 0, loc),
		time.Date(2020, time.October, 6, 1, 0, 0, 0, loc),

		"2020-10-01 01:30:00 +1030 +1030",
		"2020-10-02 01:30:00 +1030 +1030",
		"2020-10-03 01:30:00 +1030 +1030",
		"2020-10-04 01:30:00 +1030 +1030", // On 10/04 02:00 clocks go forward 30 mins.
		"2020-10-05 01:30:00 +1100 +11",
		"2020-10-06 01:30:00 +1100 +11",
	)

	// Daily & 24 hour should be identical
	check(&Rotation{
		Type:        TypeDaily,
		ShiftLength: 1,
		Start:       time.Date(2020, time.October, 2, 1, 30, 0, 0, loc),
	},
		time.Date(2020, time.October, 2, 1, 0, 0, 0, loc),
		time.Date(2020, time.October, 6, 1, 0, 0, 0, loc),

		"2020-10-01 01:30:00 +1030 +1030",
		"2020-10-02 01:30:00 +1030 +1030",
		"2020-10-03 01:30:00 +1030 +1030",
		"2020-10-04 01:30:00 +1030 +1030", // On 10/04 02:00 clocks go forward 30 mins.
		"2020-10-05 01:30:00 +1100 +11",
		"2020-10-06 01:30:00 +1100 +11",
	)

	// check when DST starts for non-CST timezone (Lord_Howe)
	// handoff time is in between the 30 min forwarding interval.
	check(&Rotation{
		Type:        TypeDaily,
		ShiftLength: 1,
		Start:       time.Date(2020, time.October, 2, 2, 15, 0, 0, loc),
	},
		time.Date(2020, time.October, 2, 1, 0, 0, 0, loc),
		time.Date(2020, time.October, 6, 1, 0, 0, 0, loc),

		"2020-10-01 02:15:00 +1030 +1030",
		"2020-10-02 02:15:00 +1030 +1030",
		"2020-10-03 02:15:00 +1030 +1030",
		"2020-10-04 02:30:00 +1100 +11", // On 10/04 02:00 clocks go forward 30 mins. Meaning 2:15 time does not actually exist anymore (since 2:00 becomes 2:30). Next immediate time that exists is 2:30.
		"2020-10-05 02:15:00 +1100 +11",
		"2020-10-06 02:15:00 +1100 +11",
	)

	// monthly rotation sanity check
	check(&Rotation{
		Type:        TypeMonthly,
		ShiftLength: 1,
		Start:       time.Date(2020, time.October, 2, 1, 30, 0, 0, loc),
	},
		time.Date(2020, time.April, 5, 1, 0, 0, 0, loc),
		time.Date(2020, time.June, 12, 1, 0, 0, 0, loc),

		"2020-04-01 00:00:00 -0500 CDT",
		"2020-05-01 00:00:00 -0500 CDT",
		"2020-06-01 00:00:00 -0500 CDT",
		"2020-07-01 00:00:00 -0500 CDT",
	)

	// check monthly rotations with shift from daylight savings to standard time
	check(&Rotation{
		Type:        TypeMonthly,
		ShiftLength: 1,
		Start:       time.Date(2020, time.October, 2, 1, 30, 0, 0, loc),
	},
		time.Date(2020, time.October, 2, 1, 0, 0, 0, loc),
		time.Date(2020, time.December, 2, 1, 0, 0, 0, loc),

		"2020-10-01 00:00:00 -0500 CDT",
		"2020-11-01 00:00:00 -0500 CDT",
		"2020-12-01 00:00:00 -0600 CST",
		"2021-01-01 00:00:00 -0600 CST",
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

	t.Run("cycle forward", func(t *testing.T) {
		// 2am june 1st
		orig := time.Date(2020, time.June, 1, 3, 0, 0, 0, time.UTC)
		rot := &Rotation{
			Type:        TypeHourly,
			ShiftLength: 10,
			Start:       orig,
		}

		res := rot.EndTime(orig.Add(30 * time.Hour))
		// 7pm following day (40 hours)
		assert.Equal(t, "2020-06-02 19:00:00 +0000 UTC", res.String())
	})
	t.Run("cycle back", func(t *testing.T) {
		// 2am june 1st
		orig := time.Date(2020, time.June, 1, 3, 0, 0, 0, time.UTC)
		rot := &Rotation{
			Type:        TypeHourly,
			ShiftLength: 10,
			Start:       orig,
		}

		res := rot.EndTime(orig.Add(-30 * time.Hour))
		// 7am previous day (-30 hours + 10 hours)
		assert.Equal(t, "2020-05-31 07:00:00 +0000 UTC", res.String())
	})

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

	// monthly
	data := []dat{
		{s: "Jun 1 2017 12:00 am", l: 1, exp: "Jul 1 2017 12:00 am", dur: time.Hour * 24 * 30},
		{s: "Jul 1 2017 12:00 am", l: 2, exp: "Sep 1 2017 12:00 am", dur: time.Hour * 24 * 31 * 2},

		// DST tests
		{s: "Mar 1 2017 12:00 am", l: 1, exp: "Apr 1 2017 12:00 am", dur: time.Hour*24*31 - time.Hour},
		{s: "Nov 1 2017 12:00 am", l: 2, exp: "Jan 1 2018 12:00 am", dur: time.Hour*(24*(30+31)) + time.Hour},
	}
	for _, d := range data {
		test(d.s, d.exp, d.l, d.dur, TypeMonthly)
	}

	// weekly
	data = []dat{
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

	t.Run("subsequent calls (hourly)", func(t *testing.T) {
		orig := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
		r := &Rotation{
			Type:        TypeHourly,
			ShiftLength: 1,
			Start:       orig,
		}
		ts := r.EndTime(orig.Add(-2 * time.Hour))
		assert.Equal(t, orig.Add(-time.Hour).String(), ts.String())

		ts = r.EndTime(ts)
		assert.Equal(t, orig.String(), ts.String())

		ts = r.EndTime(ts)
		assert.Equal(t, orig.Add(time.Hour).String(), ts.String())
	})

	t.Run("subsequent calls (daily)", func(t *testing.T) {
		orig := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
		r := &Rotation{
			Type:        TypeDaily,
			ShiftLength: 1,
			Start:       orig,
		}
		ts := r.EndTime(orig.AddDate(0, 0, -2))
		assert.Equal(t, orig.AddDate(0, 0, -1).String(), ts.String())

		ts = r.EndTime(ts)
		assert.Equal(t, orig.String(), ts.String())

		ts = r.EndTime(ts)
		assert.Equal(t, orig.AddDate(0, 0, 1).String(), ts.String())
	})

	t.Run("subsequent calls (monthly)", func(t *testing.T) {
		orig := time.Date(2020, time.January, 10, 0, 0, 0, 0, time.UTC)
		r := &Rotation{
			Type:        TypeMonthly,
			ShiftLength: 1,
			Start:       orig,
		}
		ts := r.EndTime(orig.AddDate(0, -2, 0))
		assert.Equal(t, orig.AddDate(0, -1, 0).String(), ts.String())

		ts = r.EndTime(ts)
		assert.Equal(t, orig.String(), ts.String())

		ts = r.EndTime(ts)
		assert.Equal(t, orig.AddDate(0, 1, 0).String(), ts.String())
	})
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
	require.Equal(t, time.Date(2019, 0, 4, 0, 0, 0, 0, time.UTC),
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

	// monthly
	data := []dat{
		{s: "Jun 1 2017 12:00 am", l: 1, exp: "Jul 1 2017 12:00 am", dur: time.Hour * 24 * 30},
		{s: "Jul 1 2017 12:00 am", l: 2, exp: "Sep 1 2017 12:00 am", dur: time.Hour * 24 * 31 * 2},

		// DST tests
		{s: "Mar 1 2017 12:00 am", l: 1, exp: "Apr 1 2017 12:00 am", dur: time.Hour*24*31 - time.Hour},
		{s: "Nov 1 2017 12:00 am", l: 2, exp: "Jan 1 2018 12:00 am", dur: time.Hour*(24*(31+30)) + time.Hour},
	}
	for _, d := range data {
		test(d.s, d.exp, d.l, d.dur, TypeMonthly)
	}

	// weekly
	data = []dat{
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
