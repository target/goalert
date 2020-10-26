package oncall

import (
	"sort"
	"time"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/override"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/rule"
)

type ResolvedRule struct {
	rule.Rule
	Rotation *ResolvedRotation
}
type ResolvedRotation struct {
	rotation.Rotation
	CurrentIndex int
	CurrentStart time.Time
	CurrentEnd   time.Time
	Users        []string
}

type state struct {
	tempScheds []schedule.TemporarySchedule
	rules      []ResolvedRule
	overrides  []override.UserOverride
	history    []Shift
	now        time.Time
	loc        *time.Location
}

func (r *ResolvedRotation) UserID(t time.Time) string {
	if r == nil || len(r.Users) == 0 {
		return ""
	}
	if len(r.Users) == 1 {
		return r.Users[0]
	}

	if r.CurrentStart.IsZero() {
		r.CurrentStart = r.StartTime(t)
	}

	if r.CurrentEnd.IsZero() {
		r.CurrentStart = r.StartTime(r.CurrentStart)
		r.CurrentEnd = r.EndTime(r.CurrentStart)
	}

	if t.Before(r.CurrentEnd) && !t.Before(r.CurrentStart) {
		return r.Users[r.CurrentIndex]
	}

	for !t.Before(r.CurrentEnd) {
		r.CurrentStart = r.CurrentEnd
		r.CurrentEnd = r.EndTime(r.CurrentStart)
		r.CurrentIndex++
	}
	for t.Before(r.CurrentStart) {
		r.CurrentEnd = r.CurrentStart
		r.CurrentStart = r.StartTime(r.CurrentStart.Add(-1))
		r.CurrentIndex--
	}
	r.CurrentIndex %= len(r.Users)
	if r.CurrentIndex < 0 {
		r.CurrentIndex += len(r.Users)
	}

	return r.Users[r.CurrentIndex]
}
func (r ResolvedRule) UserID(t time.Time) string {
	if !r.IsActive(t) {
		return ""
	}
	switch r.Target.TargetType() {
	case assignment.TargetTypeUser:
		return r.Target.TargetID()
	case assignment.TargetTypeRotation:
		return r.Rotation.UserID(t)
	}
	panic("unknown target type " + r.Target.TargetType().String())
}

func sortShifts(s []Shift) {
	sort.Slice(s, func(i, j int) bool {
		if s[i].Start.Equal(s[j].Start) {
			return s[i].UserID < s[j].UserID
		}
		return s[i].Start.Before(s[j].Start)
	})
}

func (s *state) CalculateShifts(start, end time.Time) []Shift {
	t := NewTimeIterator(start, end, time.Minute)
	defer t.Close()

	hist := t.NewUserCalculator()
	for _, s := range s.history {
		hist.SetSpan(s.Start, s.End, s.UserID)
	}
	hist.Init()
	tempScheds := t.NewTemporaryScheduleCalculator(s.tempScheds)
	overrides := t.NewOverrideCalculator(s.overrides)
	rules := t.NewRulesCalculator(s.loc, s.rules)

	var shifts []Shift
	isOnCall := make(map[string]*Shift)
	stillOnCall := make(map[string]bool)

	setOnCall := func(userIDs []string, startTimes []time.Time) {
		// reset map
		for id := range stillOnCall {
			delete(stillOnCall, id)
		}
		now := time.Unix(t.Unix(), 0)
		for i, id := range userIDs {
			stillOnCall[id] = true
			s := isOnCall[id]
			if s != nil {
				continue
			}
			// new user is on call
			start := now
			if len(startTimes) != 0 {
				// use original start time if provided
				start = startTimes[i]
			}
			isOnCall[id] = &Shift{
				Start:  start,
				UserID: id,
			}
		}
		for id, s := range isOnCall {
			if stillOnCall[id] {
				continue
			}

			// no longer on call
			s.End = now
			shifts = append(shifts, *s)
			delete(isOnCall, id)
		}
	}

	for t.Next() {
		if !time.Unix(t.Unix(), 0).After(s.now) {
			// use history if in the past
			setOnCall(hist.ActiveUsers(), hist.ActiveTimes())
			continue
		}

		if tempScheds.Active() {
			// use TemporarySchedule if one is active
			setOnCall(tempScheds.ActiveUsers(), nil)
			continue
		}

		// rules
		onCall := rules.ActiveUsers()

		// apply any overrides
		setOnCall(overrides.MapUsers(onCall), nil)
	}

	// remaining shifts are truncated
	for _, s := range isOnCall {
		s.Truncated = true
		s.End = time.Unix(t.Unix(), 0)
		shifts = append(shifts, *s)
	}

	sortShifts(shifts)
	return shifts
}
