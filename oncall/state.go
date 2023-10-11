package oncall

import (
	"slices"
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

func filterShiftsByUserIDs(shifts []Shift, userIDs []string) []Shift {
	var filteredShifts []Shift
	for _, shift := range shifts {
		if slices.Contains(userIDs, shift.UserID) {
			filteredShifts = append(filteredShifts, shift)
		}
	}

	return filteredShifts
}

func (s *state) CalculateShifts(start, end time.Time, userIDs []string) []Shift {
	start = start.Truncate(time.Minute)
	end = end.Truncate(time.Minute)
	tiStart := start
	if !s.now.IsZero() && s.now.Before(start) {
		tiStart = s.now.Truncate(time.Minute)
	}
	historyCutoff := s.now.Truncate(time.Minute).Add(time.Minute)
	t := NewTimeIterator(tiStart, end, time.Minute)
	defer t.Close()

	hist := t.NewUserCalculator()
	sort.Slice(s.history, func(i, j int) bool { return s.history[i].Start.Before(s.history[j].Start) })
	for _, s := range s.history {
		if s.End.IsZero() {
			// have currently active shifts "end"
			// at the cutoff.
			s.End = historyCutoff
		}
		hist.SetSpan(s.Start, s.End, s.UserID)
	}
	hist.Init()
	tempScheds := t.NewTemporaryScheduleCalculator(s.tempScheds)
	// sort overrides so that overlapping spans are merged properly

	overrides := t.NewOverrideCalculator(s.overrides)
	rules := t.NewRulesCalculator(s.loc, s.rules)

	var shifts []Shift
	isOnCall := make(map[string]*Shift)
	stillOnCall := make(map[string]bool)

	setOnCall := func(userIDs []string) {
		// reset map
		for id := range stillOnCall {
			delete(stillOnCall, id)
		}
		now := time.Unix(t.Unix(), 0)
		for _, id := range userIDs {
			stillOnCall[id] = true
			if isOnCall[id] != nil {
				continue
			}

			isOnCall[id] = &Shift{
				Start:  now,
				UserID: id,
			}
		}
		for id, shift := range isOnCall {
			if stillOnCall[id] {
				continue
			}

			// no longer on call
			if now.After(start) {
				shift.End = now
				shifts = append(shifts, *shift)
			}
			delete(isOnCall, id)
		}
	}

	for t.Next() {
		if time.Unix(t.Unix(), 0).Before(historyCutoff) {
			// use history if in the past
			setOnCall(hist.ActiveUsers())
			continue
		}

		if tempScheds.Active() {
			// use TemporarySchedule if one is active
			setOnCall(tempScheds.ActiveUsers())
			continue
		}

		// apply any overrides
		setOnCall(overrides.MapUsers(rules.ActiveUsers()))
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
