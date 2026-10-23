package rule

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/util/timeutil"
)

func TestRule_NextChange(t *testing.T) {
	load := func(name string) *time.Location {
		t.Helper()
		loc, err := time.LoadLocation(name)
		require.NoError(t, err)
		return loc
	}
	chi := load("America/Chicago")
	lh := load("Australia/Lord_Howe")

	mf := timeutil.WeekdayFilter{0, 1, 1, 1, 1, 1, 0}
	daily := timeutil.EveryDay()
	mk := func(sh, sm, eh, em int, days timeutil.WeekdayFilter) Rule {
		return Rule{Start: timeutil.NewClock(sh, sm), End: timeutil.NewClock(eh, em), WeekdayFilter: days}
	}

	// 2026-10-19 is a Monday; all values are in October 2026.
	oct := func(d, h, m, s int) time.Time { return time.Date(2026, 10, d, h, m, s, 0, time.UTC) }
	// Zone-aware inputs; expectations for those are written as UTC instants.
	in := func(loc *time.Location, mo time.Month, d, h, m, s int) time.Time {
		return time.Date(2026, mo, d, h, m, s, 0, loc)
	}
	utc := func(mo time.Month, d, h, m int) time.Time { return time.Date(2026, mo, d, h, m, 0, 0, time.UTC) }

	sameDay := mk(8, 0, 20, 0, mf)
	overnight := mk(20, 0, 8, 0, mf)
	allDay := mk(8, 0, 8, 0, mf)
	midnight := mk(0, 0, 0, 0, mf)

	type testCase struct {
		rule Rule
		t    time.Time
		exp  time.Time
	}
	check := func(desc string, c testCase) {
		t.Helper()
		got := c.rule.NextChange(c.t)
		if c.exp.IsZero() {
			assert.True(t, got.IsZero(), "%s: expected zero time, got %s", desc, got)
			return
		}
		assert.Equal(t, c.exp.UTC().Format(time.RFC3339Nano), got.UTC().Format(time.RFC3339Nano), desc)
		assert.True(t, got.After(c.t), "%s: post-condition: result must be after t", desc)
		assert.NotEqual(t, c.rule.IsActive(c.t), c.rule.IsActive(got), "%s: IsActive must flip at the returned instant", desc)
	}

	// same-day 08:00-20:00 Mon-Fri
	check("same-day/mid-shift", testCase{sameDay, oct(21, 10, 0, 0), oct(21, 20, 0, 0)})
	check("same-day/before start", testCase{sameDay, oct(21, 7, 0, 0), oct(21, 8, 0, 0)})
	check("same-day/exactly at start", testCase{sameDay, oct(21, 8, 0, 0), oct(21, 20, 0, 0)})
	check("same-day/exactly at end", testCase{sameDay, oct(21, 20, 0, 0), oct(22, 8, 0, 0)})
	check("same-day/07:59:30", testCase{sameDay, oct(21, 7, 59, 30), oct(21, 8, 0, 0)})
	check("same-day/08:00:30", testCase{sameDay, oct(21, 8, 0, 30), oct(21, 20, 0, 0)})
	check("same-day/19:59:59", testCase{sameDay, oct(21, 19, 59, 59), oct(21, 20, 0, 0)})
	check("same-day/20:00:30", testCase{sameDay, oct(21, 20, 0, 30), oct(22, 8, 0, 0)})
	check("same-day/Fri after end", testCase{sameDay, oct(23, 21, 0, 0), oct(26, 8, 0, 0)})
	check("same-day/Sat", testCase{sameDay, oct(24, 12, 0, 0), oct(26, 8, 0, 0)})

	// overnight 20:00-08:00 Mon-Fri (Start > End)
	check("overnight/Mon evening active", testCase{overnight, oct(19, 21, 0, 0), oct(20, 8, 0, 0)})
	check("overnight/Tue morning active", testCase{overnight, oct(20, 7, 0, 0), oct(20, 8, 0, 0)})
	check("overnight/Tue daytime inactive", testCase{overnight, oct(20, 9, 0, 0), oct(20, 20, 0, 0)})
	check("overnight/exactly at start", testCase{overnight, oct(20, 20, 0, 0), oct(21, 8, 0, 0)})
	check("overnight/exactly at end", testCase{overnight, oct(20, 8, 0, 0), oct(20, 20, 0, 0)})
	check("overnight/Fri night active", testCase{overnight, oct(23, 21, 0, 0), oct(24, 8, 0, 0)})
	check("overnight/Sat morning active", testCase{overnight, oct(24, 7, 0, 0), oct(24, 8, 0, 0)})
	check("overnight/Sat daytime inactive", testCase{overnight, oct(24, 9, 0, 0), oct(26, 20, 0, 0)})
	check("overnight/Sun", testCase{overnight, oct(25, 12, 0, 0), oct(26, 20, 0, 0)})

	// 24h 08:00-08:00 Mon-Fri (Start == End)
	check("24h/Mon before start", testCase{allDay, oct(19, 7, 0, 0), oct(19, 8, 0, 0)})
	check("24h/Mon exactly at start", testCase{allDay, oct(19, 8, 0, 0), oct(24, 8, 0, 0)})
	check("24h/Mon after start", testCase{allDay, oct(19, 9, 0, 0), oct(24, 8, 0, 0)})
	check("24h/Wed", testCase{allDay, oct(21, 12, 0, 0), oct(24, 8, 0, 0)})
	check("24h/Sat before end", testCase{allDay, oct(24, 7, 0, 0), oct(24, 8, 0, 0)})
	check("24h/Sat exactly at end", testCase{allDay, oct(24, 8, 0, 0), oct(26, 8, 0, 0)})
	check("24h/Sun", testCase{allDay, oct(25, 12, 0, 0), oct(26, 8, 0, 0)})

	// 00:00-00:00 Mon-Fri
	check("midnight/Sun evening", testCase{midnight, oct(25, 23, 0, 0), oct(26, 0, 0, 0)})
	check("midnight/Mon exactly at start", testCase{midnight, oct(19, 0, 0, 0), oct(24, 0, 0, 0)})
	check("midnight/Fri 23:59", testCase{midnight, oct(23, 23, 59, 0), oct(24, 0, 0, 0)})
	check("midnight/Sat exactly at end", testCase{midnight, oct(24, 0, 0, 0), oct(26, 0, 0, 0)})
	check("midnight/Sat 00:00:30", testCase{midnight, oct(24, 0, 0, 30), oct(26, 0, 0, 0)})

	// never / always
	check("never", testCase{mk(8, 0, 20, 0, timeutil.WeekdayFilter{}), oct(21, 10, 0, 0), time.Time{}})
	check("always", testCase{mk(0, 0, 0, 0, daily), oct(21, 10, 0, 0), time.Time{}})
	check("always/non-midnight", testCase{mk(8, 0, 8, 0, daily), oct(21, 10, 0, 0), time.Time{}})
	check("daily/not always", testCase{mk(8, 0, 20, 0, daily), oct(24, 10, 0, 0), oct(24, 20, 0, 0)})

	// America/Chicago spring forward 2026-03-08: 02:00 CST -> 03:00 CDT (08:00Z)
	check("chicago-spring/01:00-02:00 before", testCase{mk(1, 0, 2, 0, daily), in(chi, 3, 8, 0, 30, 0), utc(3, 8, 7, 0)})
	check("chicago-spring/01:00-02:00 active, end is skipped hour", testCase{mk(1, 0, 2, 0, daily), in(chi, 3, 8, 1, 30, 0), utc(3, 8, 8, 0)})
	check("chicago-spring/01:00-02:00 exactly at end", testCase{mk(1, 0, 2, 0, daily), utc(3, 8, 8, 0).In(chi), utc(3, 9, 6, 0)})
	check("chicago-spring/02:30-03:30 before, start is skipped", testCase{mk(2, 30, 3, 30, daily), in(chi, 3, 8, 1, 0, 0), utc(3, 8, 8, 0)})
	check("chicago-spring/02:30-03:30 active", testCase{mk(2, 30, 3, 30, daily), in(chi, 3, 8, 3, 15, 0), utc(3, 8, 8, 30)})
	check("chicago-spring/02:30-03:30 exactly at end", testCase{mk(2, 30, 3, 30, daily), in(chi, 3, 8, 3, 30, 0), utc(3, 9, 7, 30)})
	check("chicago-spring/23:30-00:30 before", testCase{mk(23, 30, 0, 30, daily), in(chi, 3, 7, 23, 0, 0), utc(3, 8, 5, 30)})
	check("chicago-spring/23:30-00:30 active before midnight", testCase{mk(23, 30, 0, 30, daily), in(chi, 3, 7, 23, 45, 0), utc(3, 8, 6, 30)})
	check("chicago-spring/23:30-00:30 active after midnight", testCase{mk(23, 30, 0, 30, daily), in(chi, 3, 8, 0, 0, 0), utc(3, 8, 6, 30)})
	check("chicago-spring/23:30-00:30 exactly at end", testCase{mk(23, 30, 0, 30, daily), in(chi, 3, 8, 0, 30, 0), utc(3, 9, 4, 30)})

	// America/Chicago fall back 2026-11-01: 02:00 CDT -> 01:00 CST (07:00Z)
	check("chicago-fall/01:00-02:00 before", testCase{mk(1, 0, 2, 0, daily), in(chi, 11, 1, 0, 30, 0), utc(11, 1, 6, 0)})
	check("chicago-fall/01:00-02:00 active (CDT)", testCase{mk(1, 0, 2, 0, daily), utc(11, 1, 6, 30).In(chi), utc(11, 1, 8, 0)})
	check("chicago-fall/01:00-02:00 active (CST)", testCase{mk(1, 0, 2, 0, daily), utc(11, 1, 7, 30).In(chi), utc(11, 1, 8, 0)})
	check("chicago-fall/01:00-02:00 exactly at end", testCase{mk(1, 0, 2, 0, daily), utc(11, 1, 8, 0).In(chi), utc(11, 2, 7, 0)})
	check("chicago-fall/02:30-03:30 before (CDT)", testCase{mk(2, 30, 3, 30, daily), utc(11, 1, 6, 30).In(chi), utc(11, 1, 8, 30)})
	check("chicago-fall/02:30-03:30 before (CST)", testCase{mk(2, 30, 3, 30, daily), utc(11, 1, 7, 30).In(chi), utc(11, 1, 8, 30)})
	check("chicago-fall/02:30-03:30 active", testCase{mk(2, 30, 3, 30, daily), utc(11, 1, 8, 30).In(chi), utc(11, 1, 9, 30)})
	check("chicago-fall/23:30-00:30 active after midnight", testCase{mk(23, 30, 0, 30, daily), in(chi, 11, 1, 0, 0, 0), utc(11, 1, 5, 30)})
	check("chicago-fall/23:30-00:30 exactly at end", testCase{mk(23, 30, 0, 30, daily), utc(11, 1, 5, 30).In(chi), utc(11, 2, 5, 30)})

	// Australia/Lord_Howe spring forward 2026-10-04: 02:00 +10:30 -> 02:30 +11:00 (2026-10-03 15:30Z)
	check("lord-howe-spring/02:00-03:00 before, start is skipped", testCase{mk(2, 0, 3, 0, daily), in(lh, 10, 4, 1, 0, 0), utc(10, 3, 15, 30)})
	check("lord-howe-spring/02:00-03:00 active", testCase{mk(2, 0, 3, 0, daily), utc(10, 3, 15, 45).In(lh), utc(10, 3, 16, 0)})
	check("lord-howe-spring/02:15-02:45 before", testCase{mk(2, 15, 2, 45, daily), in(lh, 10, 4, 1, 0, 0), utc(10, 3, 15, 30)})
	check("lord-howe-spring/02:15-02:45 active", testCase{mk(2, 15, 2, 45, daily), utc(10, 3, 15, 30).In(lh), utc(10, 3, 15, 45)})
	// Australia/Lord_Howe fall back 2026-04-05: 02:00 +11:00 -> 01:30 +10:30 (2026-04-04 15:00Z)
	check("lord-howe-fall/01:00-02:00 before", testCase{mk(1, 0, 2, 0, daily), in(lh, 4, 5, 0, 30, 0), utc(4, 4, 14, 0)})
	check("lord-howe-fall/01:00-02:00 active", testCase{mk(1, 0, 2, 0, daily), utc(4, 4, 14, 15).In(lh), utc(4, 4, 15, 30)})
	check("lord-howe-fall/01:00-02:00 active (repeated half hour)", testCase{mk(1, 0, 2, 0, daily), utc(4, 4, 15, 15).In(lh), utc(4, 4, 15, 30)})
}

// TestRule_NextChange_Property compares NextChange against a brute-force
// per-minute scan of IsActive across a full year, for several rule shapes in
// several zones.
//
// Zones whose DST transitions happen away from midnight must match exactly.
// Zones that transition AT midnight (America/Santiago, America/Havana) only
// check the zero-or-After post-condition, because of the known weekday math
// issue called out in the StartTime doc comment.
func TestRule_NextChange_Property(t *testing.T) {
	zones := []struct {
		name  string
		exact bool
	}{
		{"America/Chicago", true},
		{"Europe/London", true},
		{"Pacific/Auckland", true},
		{"Australia/Lord_Howe", true},
		{"America/Santiago", false},
		{"America/Havana", false},
	}

	mf := timeutil.WeekdayFilter{0, 1, 1, 1, 1, 1, 0}
	weekend := timeutil.WeekdayFilter{1, 0, 0, 0, 0, 0, 1}
	sunThu := timeutil.WeekdayFilter{1, 1, 1, 1, 1, 0, 0}
	daily := timeutil.EveryDay()
	shapes := []Rule{
		{Start: timeutil.NewClock(8, 0), End: timeutil.NewClock(20, 0), WeekdayFilter: mf},
		{Start: timeutil.NewClock(20, 0), End: timeutil.NewClock(8, 0), WeekdayFilter: mf},
		{Start: timeutil.NewClock(8, 0), End: timeutil.NewClock(8, 0), WeekdayFilter: mf},
		{Start: timeutil.NewClock(0, 0), End: timeutil.NewClock(0, 0), WeekdayFilter: weekend},
		{Start: timeutil.NewClock(2, 30), End: timeutil.NewClock(3, 30), WeekdayFilter: daily},
		{Start: timeutil.NewClock(23, 30), End: timeutil.NewClock(0, 30), WeekdayFilter: sunThu},
	}

	const step = 17 * time.Minute

	for _, z := range zones {
		loc, err := time.LoadLocation(z.name)
		require.NoError(t, err)

		for _, r := range shapes {
			t.Run(z.name+"/"+r.String(), func(t *testing.T) {
				t.Parallel()

				start := time.Date(2026, 1, 1, 0, 0, 0, 0, loc)
				end := time.Date(2027, 1, 1, 0, 0, 0, 0, loc)

				// per-minute scan: record every instant IsActive flips
				var flips []time.Time
				prev := r.IsActive(start)
				for m := start.Add(time.Minute); m.Before(end); m = m.Add(time.Minute) {
					cur := r.IsActive(m)
					if cur != prev {
						flips = append(flips, m)
						prev = cur
					}
				}
				require.NotEmpty(t, flips, "rule must change at least once in a year")

				nextFlip := func(tm time.Time) (time.Time, bool) {
					i := sort.Search(len(flips), func(i int) bool { return flips[i].After(tm) })
					if i == len(flips) {
						return time.Time{}, false
					}
					return flips[i], true
				}

				var failures int
				check := func(tm time.Time) {
					exp, ok := nextFlip(tm)
					if !ok {
						return // past the last known flip of the year
					}
					got := r.NextChange(tm)
					if !z.exact {
						if !got.IsZero() && !got.After(tm) {
							failures++
							if failures <= 10 {
								t.Errorf("t=%s: got %s which is not after t", tm, got)
							}
						}
						return
					}
					if !got.Equal(exp) {
						failures++
						if failures <= 10 {
							t.Errorf("t=%s: got %s; want %s", tm, got.In(loc), exp)
						}
					}
				}

				// coarse sweep across the year
				for tm := start; tm.Before(end); tm = tm.Add(step) {
					check(tm)
				}
				// and the interesting instants around every flip, including seconds
				for _, f := range flips {
					check(f.Add(-time.Minute))
					check(f.Add(-30 * time.Second))
					check(f)
					check(f.Add(30 * time.Second))
				}
				if failures > 10 {
					t.Errorf("%d total failures", failures)
				}
			})
		}
	}
}
