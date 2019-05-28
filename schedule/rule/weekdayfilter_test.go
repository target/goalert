package rule

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestWeekdayFilter_String(t *testing.T) {
	check := func(f WeekdayFilter, exp string) {
		var name strings.Builder
		for i := range f {
			if f[i] == 1 {
				name.WriteString(time.Weekday(i).String()[:1])
			} else {
				name.WriteByte('_')
			}
		}
		t.Run(name.String(), func(t *testing.T) {
			res := f.String()
			if res != exp {
				t.Errorf("got '%s'; want '%s'", res, exp)
			}
		})
	}

	check(everyDay, "every day")
	check(neverDays, "never")
	check(WeekdayFilter{1, 0, 0, 0, 0, 0, 1}, "weekends")
	check(WeekdayFilter{0, 1, 1, 1, 1, 1, 0}, "M-F")
	check(WeekdayFilter{1, 1, 0, 0, 0, 1, 0}, "Sun,Mon,Fri")
}

func TestWeekdayFilter_DaysUntil(t *testing.T) {
	check := func(f WeekdayFilter, in time.Weekday, e bool, exp int) {
		t.Run(fmt.Sprintf("%s/From%s-%t", f, in, e), func(t *testing.T) {
			res := f.DaysUntil(in, e)
			if res != exp {
				t.Errorf("got %d; want %d", res, exp)
			}
		})
	}

	check(everyDay, time.Monday, true, 0)
	check(neverDays, time.Monday, true, -1)
	check(WeekdayFilter{1, 0, 0, 0, 0, 0, 0}, time.Monday, true, 6)
	check(WeekdayFilter{0, 1, 0, 0, 0, 0, 0}, time.Monday, true, 0)
	check(WeekdayFilter{0, 0, 1, 0, 0, 0, 0}, time.Monday, true, 1)
	check(WeekdayFilter{1, 0, 1, 0, 0, 1, 0}, time.Monday, true, 1)

	check(everyDay, time.Monday, false, -1)
	check(neverDays, time.Monday, false, 0)
	check(WeekdayFilter{1, 0, 0, 0, 0, 0, 0}, time.Monday, false, 0)
	check(WeekdayFilter{0, 1, 0, 0, 0, 0, 0}, time.Monday, false, 1)
	check(WeekdayFilter{0, 0, 1, 0, 0, 0, 0}, time.Monday, false, 0)
	check(WeekdayFilter{1, 0, 1, 0, 0, 1, 0}, time.Monday, false, 0)
	check(WeekdayFilter{0, 1, 1, 1, 1, 1, 1}, time.Monday, false, 6)
}

func TestWeekdayFilter_DaysSince(t *testing.T) {
	check := func(f WeekdayFilter, in time.Weekday, e bool, exp int) {
		t.Run(fmt.Sprintf("%s/From%s-%t", f, in, e), func(t *testing.T) {
			res := f.DaysSince(in, e)
			if res != exp {
				t.Errorf("got %d; want %d", res, exp)
			}
		})
	}

	check(everyDay, time.Monday, true, 0)
	check(neverDays, time.Monday, true, -1)
	check(WeekdayFilter{1, 0, 0, 0, 0, 0, 0}, time.Monday, true, 1)
	check(WeekdayFilter{0, 1, 0, 0, 0, 0, 0}, time.Monday, true, 0)
	check(WeekdayFilter{0, 0, 1, 0, 0, 0, 0}, time.Monday, true, 6)
	check(WeekdayFilter{1, 0, 1, 0, 0, 1, 0}, time.Monday, true, 1)
	check(WeekdayFilter{0, 0, 1, 0, 0, 1, 0}, time.Monday, true, 3)
	check(WeekdayFilter{0, 0, 1, 0, 0, 1, 1}, time.Monday, true, 2)

	check(everyDay, time.Monday, false, -1)
	check(neverDays, time.Monday, false, 0)
	check(WeekdayFilter{1, 0, 0, 0, 0, 0, 0}, time.Monday, false, 0)
	check(WeekdayFilter{0, 1, 0, 0, 0, 0, 0}, time.Monday, false, 1)
	check(WeekdayFilter{0, 0, 1, 0, 0, 0, 0}, time.Monday, false, 0)
	check(WeekdayFilter{1, 0, 1, 0, 0, 1, 0}, time.Monday, false, 0)
	check(WeekdayFilter{0, 0, 1, 0, 0, 1, 0}, time.Monday, false, 0)
	check(WeekdayFilter{0, 0, 1, 0, 0, 1, 1}, time.Monday, false, 0)
	check(WeekdayFilter{1, 1, 1, 0, 0, 1, 1}, time.Monday, false, 4)

}
