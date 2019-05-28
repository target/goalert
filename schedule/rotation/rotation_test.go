package rotation

import (
	"testing"
	"time"
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

func TestRotation_EndTime_DST(t *testing.T) {
	tFmt := timeFmt + " (-07:00)"
	rot := &Rotation{
		Type:  TypeHourly,
		Start: mustParse(t, "Jan 1 2017 1:00 am"),
	}
	t.Logf("Rotation Start=%s", rot.Start.Format(tFmt))

	test := func(start, end time.Time) {
		t.Run("", func(t *testing.T) {
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
