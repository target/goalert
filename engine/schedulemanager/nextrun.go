package schedulemanager

import (
	"time"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util/timeutil"
)

// MaxLookAhead is the upper bound on how far into the future a schedule job
// will schedule its successor.
//
// Every schedule is re-evaluated at least this often even when no modeled
// boundary (rule change, temporary schedule shift, notification time, zone
// transition) falls sooner. This bounds the staleness introduced by anything
// the time math does not model: deleted users, tzdata updates, database
// restores, and the like.
const MaxLookAhead = time.Hour

// runGranularity is the grid that successor run times are aligned to.
//
// It MUST equal the River UniqueOpts.ByPeriod used when inserting the
// successor job: aligning the computed time to the same period River uses for
// uniqueness is what lets many concurrent jobs for one schedule collapse into
// a single successor.
const runGranularity = time.Minute

// ruleFromRow converts a schedule-manager rule row into a rule.Rule carrying
// only the fields relevant to time math (start, end, weekday filter).
func ruleFromRow(row gadb.SchedMgrRulesRow) rule.Rule {
	var wf timeutil.WeekdayFilter
	wf.SetDay(time.Sunday, row.Sunday)
	wf.SetDay(time.Monday, row.Monday)
	wf.SetDay(time.Tuesday, row.Tuesday)
	wf.SetDay(time.Wednesday, row.Wednesday)
	wf.SetDay(time.Thursday, row.Thursday)
	wf.SetDay(time.Friday, row.Friday)
	wf.SetDay(time.Saturday, row.Saturday)

	return rule.Rule{
		Start:         row.StartTime,
		End:           row.EndTime,
		WeekdayFilter: wf,
	}
}

// timeSpan is a half-open interval [Start, End) during which something (a
// temporary schedule, a shift, an override) is in effect. Both boundaries are
// instants at which the on-call set may change.
type timeSpan struct {
	Start, End time.Time
}

// ceilTo returns the smallest multiple of d (since the zero time) that is
// greater than or equal to t, in UTC.
func ceilTo(t time.Time, d time.Duration) time.Time {
	t = t.UTC()
	r := t.Truncate(d)
	if r.Equal(t) {
		return r
	}

	return r.Add(d)
}

// nextRunTime computes when a schedule should next be evaluated, given the
// current time and everything that is known to be able to change its state.
//
// Candidates are: each rule's next activation boundary (rule.NextChange, in
// the schedule's zone), each span's Start and End, and each notification
// time. Only candidates strictly after now are considered. The result is the
// earliest candidate, capped at the grid-aligned clamp
// now.UTC().Truncate(lookAhead).Add(lookAhead) (the next whole lookAhead
// boundary in UTC, NOT now+lookAhead, so that concurrent jobs for the same
// schedule compute the same target). If a zone transition for tz falls before
// that, it is used instead, since clock-based rules may change meaning at
// that instant. The final value is rounded up to runGranularity and returned
// in UTC.
//
// Post-conditions (hold by construction):
//
//   - result.After(now): every candidate is strictly after now, the clamp is
//     strictly after now for any positive lookAhead (Truncate <= now, then
//     +lookAhead), a zone transition lies in (now, next], and ceilTo never
//     moves a value earlier.
//   - result.Truncate(runGranularity).After(now.Truncate(runGranularity)):
//     result is already a multiple of runGranularity and is strictly after
//     now >= now.Truncate(runGranularity); two multiples of runGranularity
//     that differ must differ by at least one whole runGranularity, so the
//     bucket strictly advances. (E.g. now = 10:00:00.5 and a span end of
//     10:00:20 ceil to 10:01:00; a candidate of exactly 10:00:00 is excluded
//     by the After(now) filter rather than rounding to the current bucket.)
//     This is what prevents River from deduping the successor job against
//     the currently running one when ByPeriod == runGranularity.
//
// tz must be non-nil. A non-positive lookAhead is treated as MaxLookAhead so
// the post-conditions still hold.
func nextRunTime(now time.Time, tz *time.Location, rules []rule.Rule, spans []timeSpan, notifTimes []time.Time, lookAhead time.Duration) time.Time {
	if lookAhead <= 0 {
		lookAhead = MaxLookAhead
	}
	nowTZ := now.In(tz)

	// grid-aligned upper bound
	next := now.UTC().Truncate(lookAhead).Add(lookAhead)
	consider := func(c time.Time) {
		if c.IsZero() || !c.After(now) {
			return
		}
		if c.Before(next) {
			next = c
		}
	}

	for _, r := range rules {
		consider(r.NextChange(nowTZ))
	}
	for _, s := range spans {
		consider(s.Start)
		consider(s.End)
	}
	for _, n := range notifTimes {
		consider(n)
	}

	if tr, ok := timeutil.NextZoneTransition(nowTZ, next); ok && tr.Before(next) {
		next = tr
	}

	return ceilTo(next, runGranularity)
}
