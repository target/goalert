package rule

import (
	"testing"
	"time"
)

const timeFmt = "Mon Jan _2 3:04PM 2006"

func TestRule_IsActive(t *testing.T) {

	test := func(r Rule, tm time.Time, expected bool) {
		name := r.String() + "/" + tm.Format(timeFmt)
		if r.Start > r.End {
			name += "(overnight)"
		}
		t.Run(name, func(t *testing.T) {
			result := r.IsActive(tm)
			if result != expected {
				t.Errorf("got '%t'; want '%t'", result, expected)
			}
		})
	}

	r := Rule{
		Start: NewClock(8, 0),
		End:   NewClock(20, 0),
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
		Start: NewClock(8, 0),
		End:   NewClock(20, 0),
	}
	r.WeekdayFilter = WeekdayFilter{0, 1, 1, 1, 1, 1, 0} // M-F

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
		Start: NewClock(8, 0),
		End:   NewClock(8, 0),
	}
	r.WeekdayFilter = WeekdayFilter{0, 1, 1, 1, 1, 1, 0} // M-F

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
		Start: NewClock(20, 0),
		End:   NewClock(8, 0),
	}
	r.WeekdayFilter = WeekdayFilter{0, 1, 1, 1, 1, 1, 0} // M-F

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
		name := r.String() + "/" + start.Format(timeFmt)
		if r.Start > r.End {
			name += "(overnight)"
		}
		t.Run(name, func(t *testing.T) {
			result := r.StartTime(start)
			if !result.Equal(expected) {
				t.Errorf("got '%s'; want '%s'", result.Format(timeFmt), expected.Format(timeFmt))
			}
		})
	}

	test(Rule{
		Start:         NewClock(8, 0),
		End:           NewClock(20, 0),
		WeekdayFilter: WeekdayFilter{0, 1, 1, 0, 0, 0, 0},
	},
		time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC),
		time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC),
	)

	r := Rule{
		Start: NewClock(8, 0),
		End:   NewClock(20, 0),
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
		Start: NewClock(8, 0),
		End:   NewClock(20, 0),
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
		Start: NewClock(8, 0),
		End:   NewClock(8, 0),
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
		Start: NewClock(8, 0),
		End:   NewClock(20, 0),
	}
	r.WeekdayFilter = everyDay
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
		Start: NewClock(20, 0),
		End:   NewClock(8, 0),
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
		name := r.String() + "/" + start.Format(timeFmt)
		if r.Start > r.End {
			name += "(overnight)"
		}
		t.Run(name, func(t *testing.T) {
			result := r.EndTime(start)
			if !result.Equal(expected) {
				t.Errorf("got '%s'; want '%s'", result.Format(timeFmt), expected.Format(timeFmt))
			}
		})
	}

	r := Rule{
		Start: NewClock(8, 0),
		End:   NewClock(20, 0),
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
		Start: NewClock(8, 0),
		End:   NewClock(20, 0),
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
		Start: NewClock(8, 0),
		End:   NewClock(8, 0),
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
		Start:         NewClock(8, 0),
		End:           NewClock(20, 0),
		WeekdayFilter: everyDay,
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
		Start: NewClock(20, 0),
		End:   NewClock(8, 0),
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

	r.WeekdayFilter = everyDay

	data = []struct{ Start, Expected time.Time }{
		{Start: time.Date(2017, 7, 24, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 20, 8, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 21, 8, 0, 0, 0, time.UTC)},
		{Start: time.Date(2017, 7, 24, 20, 0, 0, 0, time.UTC), Expected: time.Date(2017, 7, 25, 8, 0, 0, 0, time.UTC)},
	}

	for _, d := range data {
		test(r, d.Start, d.Expected)
	}

}
