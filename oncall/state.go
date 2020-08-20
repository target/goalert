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
	groups    []schedule.FixedShiftGroup
	rules     []ResolvedRule
	overrides []override.UserOverride
	history   []Shift
	now       time.Time
	loc       *time.Location
}

func (r *ResolvedRotation) UserID(t time.Time) string {
	if r == nil || len(r.Users) == 0 {
		return ""
	}
	if len(r.Users) == 1 {
		return r.Users[0]
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

// trimShifts deletes and returns any shifts in the map that would not be on call at the given timestamp.
// If addNew is true:
// - Existing shifts will have their End time updated with t (if on-call at t)
// - New shifts will be added (with Start and End set to t) if on-call at t
func (s *state) trimShifts(t time.Time, m map[string]*Shift, addNew bool, appendTo []Shift) []Shift {
	active := getActiveMap()
	defer putActiveMap(active)

	ovMap := getOverrideMap()
	defer putOverrideMap(ovMap)
	for _, ov := range s.overrides {
		if !ov.End.After(t) {
			continue
		}
		if ov.Start.After(t) {
			continue
		}
		if ov.RemoveUserID == "" {
			active[ov.AddUserID] = struct{}{}
		} else {
			ovMap[ov.RemoveUserID] = ov.AddUserID
		}
	}
	for _, r := range s.rules {
		userID := r.UserID(t)
		if userID == "" {
			continue
		}

		if nextUser, ok := ovMap[userID]; ok {
			userID = nextUser
		}
		if userID == "" {
			continue
		}

		active[userID] = struct{}{}
	}

	for userID := range active {
		s, ok := m[userID]
		if !ok && addNew {
			m[userID] = &Shift{
				UserID: userID,
				Start:  t,
				End:    t,
			}
		} else if ok {
			if t.After(s.End) {
				s.End = t
			}
			if t.Before(s.Start) {
				s.Start = t
			}
		}
	}

	for userID, s := range m {
		_, ok := active[userID]
		if !ok {
			if addNew {
				s.End = t
			}
			appendTo = append(appendTo, *s)
			delete(m, userID)
		}
	}

	return appendTo
}
func (s *state) sanitize() {
	s.now = s.now.Truncate(time.Minute).In(s.loc)

	for i, o := range s.overrides {
		o.Start = o.Start.Truncate(time.Minute)
		o.End = o.End.Truncate(time.Minute)
		s.overrides[i] = o
	}
	for i, h := range s.history {
		h.End = h.End.Truncate(time.Minute)
		h.Start = h.Start.Truncate(time.Minute)
		s.history[i] = h
	}
	for _, r := range s.rules {
		if r.Rotation != nil {
			r.Rotation.Start = r.Rotation.Start.Truncate(time.Minute)
			r.Rotation.CurrentStart = r.Rotation.CurrentStart.Truncate(time.Minute)
			r.Rotation.CurrentEnd = r.Rotation.CurrentEnd.Truncate(time.Minute)
		}
	}
}
func (s *state) CalculateShifts(start, end time.Time) []Shift {
	start = start.In(s.loc).Truncate(time.Minute)
	end = end.In(s.loc).Truncate(time.Minute)
	s.sanitize()
	if !start.After(s.now) {
		start = s.now.Add(time.Minute)
	}

	curShifts := getShiftMap()
	defer putShiftMap(curShifts)

	s.trimShifts(start, curShifts, true, nil)

	shifts := make([]Shift, 0, 100)
	t := start
	for len(curShifts) > 0 && t.After(s.now) {
		t = t.Add(-time.Minute)
		shifts = s.trimShifts(t, curShifts, false, shifts)
	}

	historyShifts := make([]Shift, 0, len(s.history))
	activeShifts := make(map[string]Shift, len(s.history))
	for _, s := range s.history {
		if s.End.IsZero() {
			activeShifts[s.UserID] = s
		} else {
			historyShifts = append(historyShifts, s)
		}
	}

	// start time <= now
	for userID, s := range curShifts {
		if act, ok := activeShifts[userID]; ok {
			// We have a record of this shift, so use the provided start time.
			// If we don't, then it is not active 'now' but a rule says it will be
			// in the next minute (so we leave it as-is).
			s.Start = act.Start
		}
		shifts = append(shifts, *s)
		delete(curShifts, userID)
	}

	for _, s := range shifts {
		cpy := s
		curShifts[s.UserID] = &cpy
	}

	shifts = append(shifts[:0], historyShifts...)
	t = start
	for !t.After(end) {
		t = t.Add(time.Minute)
		shifts = s.trimShifts(t, curShifts, true, shifts)
	}
	for _, s := range curShifts { // still active @ +1 minute
		s.End = s.End.Add(-time.Minute)
		s.Truncated = true
		shifts = append(shifts, *s)
	}

	sort.Slice(shifts, func(i, j int) bool {
		if !shifts[i].Start.Equal(shifts[j].Start) {
			return shifts[i].Start.Before(shifts[j].Start)
		}
		return shifts[i].UserID < shifts[j].UserID
	})

	return shifts
}
