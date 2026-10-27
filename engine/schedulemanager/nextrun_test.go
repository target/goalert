package schedulemanager

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util/timeutil"
)

func TestRuleFromRow(t *testing.T) {
	row := gadb.SchedMgrRulesRow{
		StartTime: timeutil.NewClock(8, 0),
		EndTime:   timeutil.NewClock(20, 0),
		Monday:    true,
		Wednesday: true,
		Friday:    true,
	}
	assert.Equal(t, rule.Rule{
		Start:         timeutil.NewClock(8, 0),
		End:           timeutil.NewClock(20, 0),
		WeekdayFilter: timeutil.WeekdayFilter{0, 1, 0, 1, 0, 1, 0},
	}, ruleFromRow(row))

	row = gadb.SchedMgrRulesRow{
		StartTime: timeutil.NewClock(22, 15),
		EndTime:   timeutil.NewClock(6, 45),
		Sunday:    true,
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  true,
	}
	assert.Equal(t, rule.Rule{
		Start:         timeutil.NewClock(22, 15),
		End:           timeutil.NewClock(6, 45),
		WeekdayFilter: timeutil.EveryDay(),
	}, ruleFromRow(row))

	// ruleRowIsActive must remain equivalent to the extracted rule
	tm := time.Date(2026, 10, 21, 23, 0, 0, 0, time.UTC) // Wednesday
	assert.True(t, ruleRowIsActive(row, tm))
	assert.Equal(t, ruleFromRow(row).IsActive(tm), ruleRowIsActive(row, tm))
}

func TestCeilTo(t *testing.T) {
	chi, err := time.LoadLocation("America/Chicago")
	require.NoError(t, err)

	base := time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)
	type ceilCase struct {
		t   time.Time
		exp time.Time
	}
	check := func(desc string, c ceilCase) {
		t.Helper()
		got := ceilTo(c.t, time.Minute)
		assert.Equal(t, c.exp.Format(time.RFC3339Nano), got.Format(time.RFC3339Nano), desc)
		assert.Equal(t, time.UTC, got.Location(), desc)
		assert.False(t, got.Before(c.t), "%s: ceil must not move earlier", desc)
	}

	check("exact multiple unchanged", ceilCase{base, base})
	check("one nanosecond rounds up", ceilCase{base.Add(time.Nanosecond), base.Add(time.Minute)})
	check("20 seconds rounds up", ceilCase{base.Add(20 * time.Second), base.Add(time.Minute)})
	check("59.999s rounds up", ceilCase{base.Add(time.Minute - time.Millisecond), base.Add(time.Minute)})
	check("non-UTC input", ceilCase{base.Add(30 * time.Second).In(chi), base.Add(time.Minute)})
}

func TestNextRunTime(t *testing.T) {
	load := func(name string) *time.Location {
		t.Helper()
		loc, err := time.LoadLocation(name)
		require.NoError(t, err)
		return loc
	}
	chi := load("America/Chicago")
	stgo := load("America/Santiago")

	mf := timeutil.WeekdayFilter{0, 1, 1, 1, 1, 1, 0}
	daily := timeutil.EveryDay()
	mk := func(sh, sm, eh, em int, days timeutil.WeekdayFilter) rule.Rule {
		return rule.Rule{Start: timeutil.NewClock(sh, sm), End: timeutil.NewClock(eh, em), WeekdayFilter: days}
	}
	workHours := mk(8, 0, 20, 0, mf)
	always := rule.Rule{WeekdayFilter: daily}
	require.True(t, always.AlwaysActive())

	// 2026-10-19 is a Monday; 2026-10-21 is a Wednesday; 2026-10-23 is a Friday.
	oct := func(d, h, m, s int) time.Time { return time.Date(2026, 10, d, h, m, s, 0, time.UTC) }
	utc := func(mo time.Month, d, h, m int) time.Time { return time.Date(2026, mo, d, h, m, 0, 0, time.UTC) }
	wed10 := oct(21, 10, 0, 0)

	type nextRunCase struct {
		now       time.Time
		tz        *time.Location
		rules     []rule.Rule
		spans     []timeSpan
		notifs    []time.Time
		lookAhead time.Duration
		exp       time.Time
	}
	check := func(desc string, c nextRunCase) {
		t.Helper()
		tz := c.tz
		if tz == nil {
			tz = time.UTC
		}
		got := nextRunTime(c.now, tz, c.rules, c.spans, c.notifs, c.lookAhead)

		assert.Equal(t, c.exp.UTC().Format(time.RFC3339Nano), got.Format(time.RFC3339Nano), desc)
		assert.Equal(t, time.UTC, got.Location(), "%s: result must be in UTC", desc)
		assert.True(t, got.After(c.now), "%s: post-condition: result must be after now", desc)
		assert.True(t, got.Truncate(runGranularity).After(c.now.Truncate(runGranularity)),
			"%s: post-condition: minute bucket must strictly advance", desc)
		assert.True(t, got.Equal(got.Truncate(runGranularity)), "%s: result must be aligned to runGranularity", desc)
	}

	check("no inputs => next grid hour", nextRunCase{
		now:       oct(21, 10, 17, 33),
		lookAhead: MaxLookAhead,
		exp:       oct(21, 11, 0, 0),
	})
	check("always-active rule only => clamp", nextRunCase{
		now:       oct(21, 10, 17, 33),
		rules:     []rule.Rule{always},
		lookAhead: MaxLookAhead,
		exp:       oct(21, 11, 0, 0),
	})
	check("never-active rule only => clamp", nextRunCase{
		now:       oct(21, 10, 17, 33),
		rules:     []rule.Rule{mk(8, 0, 20, 0, timeutil.WeekdayFilter{})},
		lookAhead: MaxLookAhead,
		exp:       oct(21, 11, 0, 0),
	})
	check("M-F 08:00-20:00 at Wed 10:00 => Wed 20:00", nextRunCase{
		now:       wed10,
		rules:     []rule.Rule{workHours},
		lookAhead: 24 * time.Hour,
		exp:       oct(21, 20, 0, 0),
	})
	check("M-F 08:00-20:00 at Wed 10:00 with 1h look-ahead => clamp", nextRunCase{
		now:       wed10,
		rules:     []rule.Rule{workHours},
		lookAhead: MaxLookAhead,
		exp:       oct(21, 11, 0, 0),
	})
	check("multiple rules => earliest boundary", nextRunCase{
		now:       wed10,
		rules:     []rule.Rule{workHours, mk(11, 0, 12, 0, daily)},
		lookAhead: 24 * time.Hour,
		exp:       oct(21, 11, 0, 0),
	})
	check("Fri 21:00 with 1h look-ahead => clamp", nextRunCase{
		now:       oct(23, 21, 0, 0),
		rules:     []rule.Rule{workHours},
		lookAhead: MaxLookAhead,
		exp:       oct(23, 22, 0, 0),
	})
	// The clamp is grid-aligned (next multiple of lookAhead since the
	// zero time), NOT now+lookAhead: the 72h grid happens to land on
	// Mon 00:00Z here, which precedes the Mon 08:00 rule boundary.
	check("Fri 21:00 with 72h look-ahead => grid clamp Mon 00:00", nextRunCase{
		now:       oct(23, 21, 0, 0),
		rules:     []rule.Rule{workHours},
		lookAhead: 72 * time.Hour,
		exp:       oct(26, 0, 0, 0),
	})
	// 120h grid from Fri 21:00Z lands on Tue 00:00Z, after Mon 08:00.
	check("Fri 21:00 with 120h look-ahead => Mon 08:00", nextRunCase{
		now:       oct(23, 21, 0, 0),
		rules:     []rule.Rule{workHours},
		lookAhead: 120 * time.Hour,
		exp:       oct(26, 8, 0, 0),
	})
	check("future span start sooner than rule boundary", nextRunCase{
		now:       wed10,
		rules:     []rule.Rule{workHours},
		spans:     []timeSpan{{Start: oct(21, 12, 0, 0), End: oct(22, 12, 0, 0)}},
		lookAhead: 24 * time.Hour,
		exp:       oct(21, 12, 0, 0),
	})
	check("active span end", nextRunCase{
		now:       wed10,
		rules:     []rule.Rule{workHours},
		spans:     []timeSpan{{Start: oct(21, 9, 0, 0), End: oct(21, 11, 30, 0)}},
		lookAhead: 24 * time.Hour,
		exp:       oct(21, 11, 30, 0),
	})
	check("span ending exactly at now excluded", nextRunCase{
		now:       wed10,
		spans:     []timeSpan{{Start: oct(21, 9, 0, 0), End: wed10}},
		lookAhead: MaxLookAhead,
		exp:       oct(21, 11, 0, 0),
	})
	check("span in the past excluded", nextRunCase{
		now:       wed10,
		spans:     []timeSpan{{Start: oct(20, 9, 0, 0), End: oct(20, 11, 0, 0)}},
		lookAhead: MaxLookAhead,
		exp:       oct(21, 11, 0, 0),
	})
	check("zero span boundary ignored", nextRunCase{
		now:       wed10,
		spans:     []timeSpan{{End: oct(21, 12, 0, 0)}},
		lookAhead: 24 * time.Hour,
		exp:       oct(21, 12, 0, 0),
	})
	check("span end 10:00:20 with now 10:00:00 => 10:01:00", nextRunCase{
		now:       wed10,
		spans:     []timeSpan{{Start: oct(21, 9, 0, 0), End: oct(21, 10, 0, 20)}},
		lookAhead: MaxLookAhead,
		exp:       oct(21, 10, 1, 0),
	})
	check("sub-second now with span end in the same minute => next minute", nextRunCase{
		now:       wed10.Add(500 * time.Millisecond),
		spans:     []timeSpan{{Start: oct(21, 9, 0, 0), End: oct(21, 10, 0, 20)}},
		lookAhead: MaxLookAhead,
		exp:       oct(21, 10, 1, 0),
	})
	check("notif time sooner than everything", nextRunCase{
		now:       wed10,
		rules:     []rule.Rule{workHours},
		spans:     []timeSpan{{Start: oct(21, 12, 0, 0), End: oct(22, 12, 0, 0)}},
		notifs:    []time.Time{oct(21, 10, 5, 0)},
		lookAhead: 24 * time.Hour,
		exp:       oct(21, 10, 5, 0),
	})
	check("notif time not after now excluded", nextRunCase{
		now:       wed10,
		notifs:    []time.Time{wed10, oct(21, 9, 0, 0)},
		lookAhead: MaxLookAhead,
		exp:       oct(21, 11, 0, 0),
	})
	check("zoned rule evaluated in schedule zone", nextRunCase{
		now:       wed10.In(chi), // 05:00 CDT
		tz:        chi,
		rules:     []rule.Rule{workHours},
		lookAhead: 24 * time.Hour,
		exp:       oct(21, 13, 0, 0), // 08:00 CDT
	})
	check("Chicago 02:30-03:30 rule at 01:00 CST on 2026-03-08 => 03:00 CDT", nextRunCase{
		now:       time.Date(2026, 3, 8, 1, 0, 0, 0, chi),
		tz:        chi,
		rules:     []rule.Rule{mk(2, 30, 3, 30, daily)},
		lookAhead: 24 * time.Hour,
		exp:       utc(time.March, 8, 8, 0),
	})
	check("Chicago zone transition pre-empts a later clamp", nextRunCase{
		now:       time.Date(2026, 3, 8, 1, 0, 0, 0, chi),
		tz:        chi,
		lookAhead: 24 * time.Hour,
		exp:       utc(time.March, 8, 8, 0),
	})
	check("Chicago zone transition pre-empts a later rule boundary", nextRunCase{
		now:       time.Date(2026, 3, 8, 1, 0, 0, 0, chi),
		tz:        chi,
		rules:     []rule.Rule{workHours}, // Sunday: next boundary is Mon 08:00
		lookAhead: 24 * time.Hour,
		exp:       utc(time.March, 8, 8, 0),
	})
	check("Chicago fall back transition found", nextRunCase{
		now:       time.Date(2026, 11, 1, 0, 30, 0, 0, chi), // 05:30Z
		tz:        chi,
		lookAhead: 24 * time.Hour,
		exp:       utc(time.November, 1, 7, 0),
	})
	// Sat 2026-09-05 20:30 -04 is 00:30Z; clamp 01:00Z; the midnight
	// transition (04:00Z) is beyond the clamp => min(clamp, transition).
	check("Santiago Sat 20:30 local with 1h look-ahead => clamp", nextRunCase{
		now:       time.Date(2026, 9, 5, 20, 30, 0, 0, stgo),
		tz:        stgo,
		rules:     []rule.Rule{mk(8, 0, 20, 0, daily)},
		lookAhead: MaxLookAhead,
		exp:       utc(time.September, 6, 1, 0),
	})
	// clamp 06:00Z; the midnight transition at 04:00Z wins.
	check("Santiago Sat 20:30 local with 6h look-ahead => midnight transition", nextRunCase{
		now:       time.Date(2026, 9, 5, 20, 30, 0, 0, stgo),
		tz:        stgo,
		rules:     []rule.Rule{mk(8, 0, 20, 0, daily)},
		lookAhead: 6 * time.Hour,
		exp:       utc(time.September, 6, 4, 0),
	})
	check("grid clamp: start of hour", nextRunCase{
		now:       wed10.Add(500 * time.Millisecond),
		lookAhead: MaxLookAhead,
		exp:       oct(21, 11, 0, 0),
	})
	check("grid clamp: end of hour", nextRunCase{
		now:       oct(21, 10, 59, 59),
		lookAhead: MaxLookAhead,
		exp:       oct(21, 11, 0, 0),
	})
	check("grid clamp: exactly on the hour", nextRunCase{
		now:       wed10,
		lookAhead: MaxLookAhead,
		exp:       oct(21, 11, 0, 0),
	})
	check("non-positive look-ahead falls back to MaxLookAhead", nextRunCase{
		now:       oct(21, 10, 17, 33),
		lookAhead: 0,
		exp:       oct(21, 11, 0, 0),
	})
}
