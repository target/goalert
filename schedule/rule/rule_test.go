package rule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/util/timeutil"
)

const timeFmt = "Mon Jan _2 3:04PM 2006"

func TestRule_IsActive(t *testing.T) {

	test := func(r Rule, tm time.Time, expected bool) {
		t.Helper()
		name := r.String() + "/" + tm.Format(timeFmt)
		if r.Start > r.End {
			name += "(overnight)"
		}
		t.Run(name, func(t *testing.T) {
			t.Helper()
			result := r.IsActive(tm)
			if result != expected {
				t.Errorf("got '%t'; want '%t'", result, expected)
			}
		})
	}

	r := Rule{
		Start: timeutil.NewClock(8, 0),
		End:   timeutil.NewClock(20, 0),
	}
	r.SetDay(time.Monday, true)

	data := []struct {
		Time   time.Time
		Active bool
	}{
		{Time: time.Date(2017, 7, 20, 8, 0, 0, 0, time.UTC), Active: false},  // before
		{Time: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC), Active: false},  // after
		{Time: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC), Active: true},   // eq start
		{Time: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC), Active: false}, // eq end
		{Time: time.Date(2017, 7, 24, 9, 0, 0, 0, time.UTC), Active: true},   // middle
	}

	for _, d := range data {
		test(r, d.Time, d.Active)
	}
	// overnight
	r.Start, r.End = r.End, r.Start
	data = []struct {
		Time   time.Time
		Active bool
	}{
		{Time: time.Date(2017, 7, 20, 8, 0, 0, 0, time.UTC), Active: false}, // before
		{Time: time.Date(2017, 7, 26, 8, 0, 0, 0, time.UTC), Active: false}, // after
		{Time: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC), Active: true}, // eq start
		{Time: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC), Active: false}, // eq end
		{Time: time.Date(2017, 7, 24, 9, 0, 0, 0, time.UTC), Active: false}, // middle (wrong side)
		{Time: time.Date(2017, 7, 24, 21, 0, 0, 0, time.UTC), Active: true}, // middle
		{Time: time.Date(2017, 7, 25, 7, 0, 0, 0, time.UTC), Active: true},  // middle (next day)
	}
	for _, d := range data {
		test(r, d.Time, d.Active)
	}

	// weekday filters
	r = Rule{
		Start: timeutil.NewClock(8, 0),
		End:   timeutil.NewClock(20, 0),
	}
	r.WeekdayFilter = timeutil.WeekdayFilter{0, 1, 1, 1, 1, 1, 0} // M-F

	data = []struct {
		Time   time.Time
		Active bool
	}{
		{Time: time.Date(2017, 7, 23, 8, 0, 0, 0, time.UTC), Active: false}, // sun
		{Time: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC), Active: true},  // mon
		{Time: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC), Active: true},  // tues
		{Time: time.Date(2017, 7, 26, 8, 0, 0, 0, time.UTC), Active: true},  // wed
		{Time: time.Date(2017, 7, 27, 8, 0, 0, 0, time.UTC), Active: true},  // thurs
		{Time: time.Date(2017, 7, 28, 8, 0, 0, 0, time.UTC), Active: true},  // fri
		{Time: time.Date(2017, 7, 29, 8, 0, 0, 0, time.UTC), Active: false}, // sat
	}
	for _, d := range data {
		test(r, d.Time, d.Active)
	}

	// contig. filters
	r = Rule{
		Start: timeutil.NewClock(8, 0),
		End:   timeutil.NewClock(8, 0),
	}
	r.WeekdayFilter = timeutil.WeekdayFilter{0, 1, 1, 1, 1, 1, 0} // M-F

	data = []struct {
		Time   time.Time
		Active bool
	}{
		{Time: time.Date(2017, 7, 23, 8, 0, 0, 0, time.UTC), Active: false}, // sun
		{Time: time.Date(2017, 7, 24, 7, 0, 0, 0, time.UTC), Active: false}, // mon (morn)
		{Time: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC), Active: true},  // mon
		{Time: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC), Active: true},  // tues
		{Time: time.Date(2017, 7, 26, 8, 0, 0, 0, time.UTC), Active: true},  // wed
		{Time: time.Date(2017, 7, 27, 8, 0, 0, 0, time.UTC), Active: true},  // thurs
		{Time: time.Date(2017, 7, 28, 8, 0, 0, 0, time.UTC), Active: true},  // fri
		{Time: time.Date(2017, 7, 29, 7, 0, 0, 0, time.UTC), Active: true},  // sat (morn)
		{Time: time.Date(2017, 7, 29, 8, 0, 0, 0, time.UTC), Active: false}, // sat
	}
	for _, d := range data {
		test(r, d.Time, d.Active)
	}

	// weekday overnight
	r = Rule{
		Start: timeutil.NewClock(20, 0),
		End:   timeutil.NewClock(8, 0),
	}
	r.WeekdayFilter = timeutil.WeekdayFilter{0, 1, 1, 1, 1, 1, 0} // M-F

	data = []struct {
		Time   time.Time
		Active bool
	}{
		{Time: time.Date(2017, 7, 23, 20, 0, 0, 0, time.UTC), Active: false}, // sun-mon
		{Time: time.Date(2017, 7, 24, 7, 0, 0, 0, time.UTC), Active: false},

		{Time: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC), Active: true}, // mon-tues
		{Time: time.Date(2017, 7, 25, 7, 0, 0, 0, time.UTC), Active: true},

		{Time: time.Date(2017, 7, 25, 20, 0, 0, 0, time.UTC), Active: true}, // tues-wed
		{Time: time.Date(2017, 7, 26, 7, 0, 0, 0, time.UTC), Active: true},

		{Time: time.Date(2017, 7, 26, 20, 0, 0, 0, time.UTC), Active: true}, // wed-thurs
		{Time: time.Date(2017, 7, 27, 7, 0, 0, 0, time.UTC), Active: true},

		{Time: time.Date(2017, 7, 27, 20, 0, 0, 0, time.UTC), Active: true}, // thurs-fri
		{Time: time.Date(2017, 7, 28, 7, 0, 0, 0, time.UTC), Active: true},

		{Time: time.Date(2017, 7, 28, 20, 0, 0, 0, time.UTC), Active: true}, // fri-sat
		{Time: time.Date(2017, 7, 29, 7, 0, 0, 0, time.UTC), Active: true},

		{Time: time.Date(2017, 7, 29, 20, 0, 0, 0, time.UTC), Active: false}, // sat-sun
		{Time: time.Date(2017, 7, 30, 7, 0, 0, 0, time.UTC), Active: false},
	}

	for _, d := range data {
		test(r, d.Time, d.Active)
	}
}

func TestRule_StartTime(t *testing.T) {

	test := func(r Rule, start, expected time.Time) {
		t.Helper()
		name := r.String() + "/" + start.Format(timeFmt)
		if r.Start > r.End {
			name += "(overnight)"
		}
		t.Run(name, func(t *testing.T) {
			t.Helper()
			result := r.StartTime(start)
			if !result.Equal(expected) {
				t.Errorf("got '%s'; want '%s'", result.Format(timeFmt), expected.Format(timeFmt))
			}
		})
	}

	test(Rule{
		Start:         timeutil.NewClock(8, 0),
		End:           timeutil.NewClock(20, 0),
		WeekdayFilter: timeutil.WeekdayFilter{0, 1, 1, 0, 0, 0, 0},
	},
		time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC),
		time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC),
	)

	r := Rule{
		Start: timeutil.NewClock(8, 0),
		End:   timeutil.NewClock(20, 0),
	}
	r.SetDay(time.Monday, true)
	// jul 24 2017 is a Monday

	data := []struct{ Start, Expected time.Time }{

		// should be next monday shift
		{Start: time.Date(2017, 7, 20, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},

		{Start: time.Date(2017, 7, 24, 7, 59, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 8, 1, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},

		{Start: time.Date(2017, 7, 24, 19, 59, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},

		// following monday
		{Start: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 31, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 20, 1, 0, 0, time.UTC), Expected: time.Date(2017, 7, 31, 8, 0, 0, 0, time.UTC)},
	}
	for _, d := range data {
		test(r, d.Start, d.Expected)
	}

	r = Rule{
		Start: timeutil.NewClock(8, 0),
		End:   timeutil.NewClock(20, 0),
	}
	r.SetDay(time.Friday, true)
	r.SetDay(time.Saturday, true)
	r.SetDay(time.Monday, true)

	data = []struct{ Start, Expected time.Time }{
		// should be next monday shift
		{Start: time.Date(2017, 7, 21, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 22, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 22, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},

		{Start: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 28, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 25, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 28, 8, 0, 0, 0, time.UTC)},
	}
	for _, d := range data {
		test(r, d.Start, d.Expected)
	}

	r = Rule{
		Start: timeutil.NewClock(8, 0),
		End:   timeutil.NewClock(8, 0),
	}
	r.SetDay(time.Monday, true)
	r.SetDay(time.Tuesday, true)

	data = []struct{ Start, Expected time.Time }{
		// should be next monday shift
		{Start: time.Date(2017, 7, 20, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},

		{Start: time.Date(2017, 7, 24, 7, 59, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 8, 1, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},

		{Start: time.Date(2017, 7, 24, 19, 59, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},

		// following monday
		{Start: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 20, 1, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},

		{Start: time.Date(2017, 7, 25, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 25, 20, 1, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 26, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 31, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 26, 20, 1, 0, 0, time.UTC), Expected: time.Date(2017, 7, 31, 8, 0, 0, 0, time.UTC)},
	}
	for _, d := range data {
		test(r, d.Start, d.Expected)
	}

	r = Rule{
		Start: timeutil.NewClock(8, 0),
		End:   timeutil.NewClock(20, 0),
	}
	r.WeekdayFilter = timeutil.EveryDay()
	data = []struct{ Start, Expected time.Time }{

		{Start: time.Date(2017, 7, 20, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 20, 8, 0, 0, 0, time.UTC)},

		{Start: time.Date(2017, 7, 24, 7, 59, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 8, 1, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},

		{Start: time.Date(2017, 7, 24, 19, 59, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC)},

		// following monday
		{Start: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 20, 1, 0, 0, time.UTC), Expected: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC)},
	}
	for _, d := range data {
		test(r, d.Start, d.Expected)
	}

	// overnight
	r = Rule{
		Start: timeutil.NewClock(20, 0),
		End:   timeutil.NewClock(8, 0),
	}
	r.SetDay(time.Monday, true)
	// July 24th is a Monday

	data = []struct{ Start, Expected time.Time }{
		{Start: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 25, 7, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC)},

		{Start: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 31, 20, 0, 0, 0, time.UTC)},
	}

	for _, d := range data {
		test(r, d.Start, d.Expected)
	}

}

func TestRule_EndTime(t *testing.T) {
	test := func(r Rule, start, expected time.Time) {
		t.Helper()
		name := r.String() + "/" + start.Format(timeFmt)
		if r.Start > r.End {
			name += "(overnight)"
		}
		t.Run(name, func(t *testing.T) {
			t.Helper()
			result := r.EndTime(start)
			if !result.Equal(expected) {
				t.Errorf("got '%s'; want '%s'", result.Format(timeFmt), expected.Format(timeFmt))
			}
		})
	}

	r := Rule{
		Start: timeutil.NewClock(8, 0),
		End:   timeutil.NewClock(20, 0),
	}
	r.SetDay(time.Monday, true)

	data := []struct{ Start, Expected time.Time }{
		{Start: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 20, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 31, 20, 0, 0, 0, time.UTC)},
	}
	for _, d := range data {
		test(r, d.Start, d.Expected)
	}

	r = Rule{
		Start: timeutil.NewClock(8, 0),
		End:   timeutil.NewClock(20, 0),
	}
	r.SetDay(time.Monday, true)
	r.SetDay(time.Tuesday, true)

	data = []struct{ Start, Expected time.Time }{
		{Start: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 20, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 25, 20, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 25, 20, 0, 0, 0, time.UTC)},
	}
	for _, d := range data {
		test(r, d.Start, d.Expected)
	}

	r = Rule{
		Start: timeutil.NewClock(8, 0),
		End:   timeutil.NewClock(8, 0),
	}
	r.SetDay(time.Monday, true)
	r.SetDay(time.Tuesday, true)

	data = []struct{ Start, Expected time.Time }{
		{Start: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 26, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 20, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 26, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 26, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 26, 8, 0, 0, 0, time.UTC)},
	}
	for _, d := range data {
		test(r, d.Start, d.Expected)
	}

	r = Rule{
		Start:         timeutil.NewClock(8, 0),
		End:           timeutil.NewClock(20, 0),
		WeekdayFilter: timeutil.EveryDay(),
	}

	data = []struct{ Start, Expected time.Time }{
		{Start: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 20, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 20, 20, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 25, 20, 0, 0, 0, time.UTC)},
	}

	for _, d := range data {
		test(r, d.Start, d.Expected)
	}

	// overnight
	r = Rule{
		Start: timeutil.NewClock(20, 0),
		End:   timeutil.NewClock(8, 0),
	}
	r.SetDay(time.Monday, true)

	data = []struct{ Start, Expected time.Time }{
		{Start: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 8, 1, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 20, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC)},
	}

	for _, d := range data {
		test(r, d.Start, d.Expected)
	}

	r.WeekdayFilter = timeutil.EveryDay()

	data = []struct{ Start, Expected time.Time }{
		{Start: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 20, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 21, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC)},
	}

	for _, d := range data {
		test(r, d.Start, d.Expected)
	}

}

func TestRule_MidnightDST(t *testing.T) {
	check := func(r Rule, ts time.Time, expActive bool, expStart, expEnd string) {
		t.Helper()
		t.Run(r.String()+"/"+ts.String(), func(t *testing.T) {
			assert.Equal(t, expActive, r.IsActive(ts), "IsActive")
			assert.Equal(t, expStart, r.StartTime(ts).String(), "StartTime")
			assert.Equal(t, expEnd, r.EndTime(ts).String(), "EndTime")
		})
	}

	day := Rule{Start: timeutil.NewClock(8, 0), End: timeutil.NewClock(20, 0), WeekdayFilter: timeutil.EveryDay()}
	night := Rule{Start: timeutil.NewClock(20, 0), End: timeutil.NewClock(8, 0), WeekdayFilter: timeutil.EveryDay()}

	// DST starts at midnight in Santiago on Sunday 2026-09-06 (00:00 -> 01:00),
	// so local midnight does not exist and the day starts at 01:00 -03.
	santiago, err := time.LoadLocation("America/Santiago")
	require.NoError(t, err)
	at := func(d, h, m int) time.Time { return time.Date(2026, time.September, d, h, m, 0, 0, santiago) }

	check(day, at(5, 19, 59), true, "2026-09-05 08:00:00 -0400 -04", "2026-09-05 20:00:00 -0400 -04")
	check(day, at(5, 20, 0), false, "2026-09-06 08:00:00 -0300 -03", "2026-09-06 20:00:00 -0300 -03")
	check(day, at(5, 20, 30), false, "2026-09-06 08:00:00 -0300 -03", "2026-09-06 20:00:00 -0300 -03")
	check(day, at(5, 21, 0), false, "2026-09-06 08:00:00 -0300 -03", "2026-09-06 20:00:00 -0300 -03")
	check(day, at(5, 23, 59), false, "2026-09-06 08:00:00 -0300 -03", "2026-09-06 20:00:00 -0300 -03")
	check(day, at(6, 1, 0), false, "2026-09-06 08:00:00 -0300 -03", "2026-09-06 20:00:00 -0300 -03")
	check(day, at(6, 7, 59), false, "2026-09-06 08:00:00 -0300 -03", "2026-09-06 20:00:00 -0300 -03")
	check(day, at(6, 8, 0), true, "2026-09-06 08:00:00 -0300 -03", "2026-09-06 20:00:00 -0300 -03")
	check(day, at(6, 19, 59), true, "2026-09-06 08:00:00 -0300 -03", "2026-09-06 20:00:00 -0300 -03")
	check(day, at(6, 20, 0), false, "2026-09-07 08:00:00 -0300 -03", "2026-09-07 20:00:00 -0300 -03")
	check(day, at(7, 8, 0), true, "2026-09-07 08:00:00 -0300 -03", "2026-09-07 20:00:00 -0300 -03")

	check(night, at(5, 19, 59), false, "2026-09-05 20:00:00 -0400 -04", "2026-09-06 08:00:00 -0300 -03")
	check(night, at(5, 20, 0), true, "2026-09-05 20:00:00 -0400 -04", "2026-09-06 08:00:00 -0300 -03")
	check(night, at(5, 23, 59), true, "2026-09-05 20:00:00 -0400 -04", "2026-09-06 08:00:00 -0300 -03")
	check(night, at(6, 1, 0), true, "2026-09-05 20:00:00 -0400 -04", "2026-09-06 08:00:00 -0300 -03")
	check(night, at(6, 7, 59), true, "2026-09-05 20:00:00 -0400 -04", "2026-09-06 08:00:00 -0300 -03")
	check(night, at(6, 8, 0), false, "2026-09-06 20:00:00 -0300 -03", "2026-09-07 08:00:00 -0300 -03")
	check(night, at(6, 20, 0), true, "2026-09-06 20:00:00 -0300 -03", "2026-09-07 08:00:00 -0300 -03")
	check(night, at(7, 0, 30), true, "2026-09-06 20:00:00 -0300 -03", "2026-09-07 08:00:00 -0300 -03") // 00:30 the day before does not exist
	check(night, at(7, 7, 59), true, "2026-09-06 20:00:00 -0300 -03", "2026-09-07 08:00:00 -0300 -03")

	// 00:30 does not exist on the 6th; the first instant after (01:00) is used in its place
	late := Rule{Start: timeutil.NewClock(23, 0), End: timeutil.NewClock(0, 30), WeekdayFilter: timeutil.EveryDay()}
	check(late, at(5, 23, 30), true, "2026-09-05 23:00:00 -0400 -04", "2026-09-06 01:00:00 -0300 -03")
	check(late, at(6, 1, 0), false, "2026-09-06 23:00:00 -0300 -03", "2026-09-07 00:30:00 -0300 -03")

	early := Rule{Start: timeutil.NewClock(0, 30), End: timeutil.NewClock(8, 0), WeekdayFilter: timeutil.EveryDay()}
	check(early, at(5, 23, 0), false, "2026-09-06 01:00:00 -0300 -03", "2026-09-06 08:00:00 -0300 -03")
	check(early, at(6, 1, 0), true, "2026-09-06 01:00:00 -0300 -03", "2026-09-06 08:00:00 -0300 -03")
	check(early, at(6, 7, 59), true, "2026-09-06 01:00:00 -0300 -03", "2026-09-06 08:00:00 -0300 -03")
	check(early, at(6, 8, 0), false, "2026-09-07 00:30:00 -0300 -03", "2026-09-07 08:00:00 -0300 -03")

	// weekday filters
	sunOnly := Rule{Start: timeutil.NewClock(8, 0), End: timeutil.NewClock(20, 0), WeekdayFilter: timeutil.WeekdayFilter{1, 0, 0, 0, 0, 0, 0}}
	check(sunOnly, at(4, 10, 0), false, "2026-09-06 08:00:00 -0300 -03", "2026-09-06 20:00:00 -0300 -03")
	check(sunOnly, at(6, 10, 0), true, "2026-09-06 08:00:00 -0300 -03", "2026-09-06 20:00:00 -0300 -03")
	check(sunOnly, at(6, 20, 0), false, "2026-09-13 08:00:00 -0300 -03", "2026-09-13 20:00:00 -0300 -03")

	sunAllDay := Rule{Start: timeutil.NewClock(0, 0), End: timeutil.NewClock(0, 0), WeekdayFilter: timeutil.WeekdayFilter{1, 0, 0, 0, 0, 0, 0}}
	check(sunAllDay, at(5, 23, 59), false, "2026-09-06 01:00:00 -0300 -03", "2026-09-07 00:00:00 -0300 -03")
	check(sunAllDay, at(6, 1, 0), true, "2026-09-06 01:00:00 -0300 -03", "2026-09-07 00:00:00 -0300 -03")
	check(sunAllDay, at(6, 23, 59), true, "2026-09-06 01:00:00 -0300 -03", "2026-09-07 00:00:00 -0300 -03")
	check(sunAllDay, at(7, 0, 0), false, "2026-09-13 00:00:00 -0300 -03", "2026-09-14 00:00:00 -0300 -03")

	// DST starts at midnight in Havana on Sunday 2026-03-08 (00:00 -> 01:00)
	havana, err := time.LoadLocation("America/Havana")
	require.NoError(t, err)
	at = func(d, h, m int) time.Time { return time.Date(2026, time.March, d, h, m, 0, 0, havana) }

	check(day, at(7, 20, 30), false, "2026-03-08 08:00:00 -0400 CDT", "2026-03-08 20:00:00 -0400 CDT")
	check(day, at(8, 1, 0), false, "2026-03-08 08:00:00 -0400 CDT", "2026-03-08 20:00:00 -0400 CDT")
	check(day, at(8, 8, 0), true, "2026-03-08 08:00:00 -0400 CDT", "2026-03-08 20:00:00 -0400 CDT")
	check(night, at(8, 1, 0), true, "2026-03-07 20:00:00 -0500 CST", "2026-03-08 08:00:00 -0400 CDT")
	check(night, at(9, 0, 30), true, "2026-03-08 20:00:00 -0400 CDT", "2026-03-09 08:00:00 -0400 CDT")

	// DST ends in Havana on Sunday 2026-11-01 (01:00 -> 00:00), so midnight happens twice
	utc := func(d, h, m int) time.Time { return time.Date(2026, time.November, d, h, m, 0, 0, time.UTC).In(havana) }
	check(day, utc(1, 0, 30), false, "2026-11-01 08:00:00 -0500 CST", "2026-11-01 20:00:00 -0500 CST")   // Oct 31 20:30 CDT
	check(day, utc(1, 4, 30), false, "2026-11-01 08:00:00 -0500 CST", "2026-11-01 20:00:00 -0500 CST")   // 00:30 CDT
	check(day, utc(1, 5, 30), false, "2026-11-01 08:00:00 -0500 CST", "2026-11-01 20:00:00 -0500 CST")   // 00:30 CST
	check(day, utc(1, 13, 0), true, "2026-11-01 08:00:00 -0500 CST", "2026-11-01 20:00:00 -0500 CST")    // 08:00 CST
	check(night, utc(1, 4, 30), true, "2026-10-31 20:00:00 -0400 CDT", "2026-11-01 08:00:00 -0500 CST")  // 00:30 CDT
	check(night, utc(1, 5, 30), true, "2026-10-31 20:00:00 -0400 CDT", "2026-11-01 08:00:00 -0500 CST")  // 00:30 CST
	check(night, utc(1, 13, 0), false, "2026-11-01 20:00:00 -0500 CST", "2026-11-02 08:00:00 -0500 CST") // 08:00 CST
}
