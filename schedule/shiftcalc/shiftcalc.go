package shiftcalc

import (
	"context"
	"fmt"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/override"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util/log"
	"sort"
	"time"

	"github.com/pkg/errors"
)

type Calculator interface {
	RotationShifts(ctx context.Context, start, end time.Time, rotationID string) ([]RotationShift, error)
	ScheduleAssignments(ctx context.Context, start, end time.Time, scheduleID string) ([]ScheduleAssignment, error)
	ScheduleFinalShifts(ctx context.Context, start, end time.Time, scheduleID string) ([]Shift, error)
	ScheduleFinalShiftsWithOverrides(ctx context.Context, start, end time.Time, scheduleID string) ([]Shift, error)
}
type Shift struct {
	Start  time.Time `json:"start_time"`
	End    time.Time `json:"end_time"`
	UserID string    `json:"user_id"`
}

const debugTimeFmt = "MonJan2_2006@3:04pm"

func (s Shift) String() string {
	return fmt.Sprintf("Shift{Start: %s, End: %s, UserID: %s}",
		s.Start.Local().Format(debugTimeFmt),
		s.End.Local().Format(debugTimeFmt),
		s.UserID,
	)
}

type ShiftCalculator struct {
	RuleStore  rule.Store
	SchedStore schedule.Store
	RotStore   rotation.Store
	OStore     override.Store
}

type ScheduleAssignment struct {
	Target     assignment.Target `json:"target"`
	ScheduleID string            `json:"schedule_id"`
	Rules      []rule.Rule       `json:"rules"`
	Shifts     []Shift           `json:"shifts"`
}

type RotationShift struct {
	Start  time.Time `json:"start_time"`
	End    time.Time `json:"end_time"`
	PartID string    `json:"participant_id"`
}

type data struct {
	sched         schedule.Schedule
	rules         []rule.Rule
	rots          []rotation.Rotation
	parts         []rotation.Participant
	rState        []rotation.State
	userOverrides []override.UserOverride
}

func (d *data) rulesByTarget() map[assignment.RawTarget][]rule.Rule {
	m := make(map[assignment.RawTarget][]rule.Rule, len(d.rules))
	for _, r := range d.rules {
		raw := assignment.NewRawTarget(r.Target)
		m[raw] = append(m[raw], r)
	}
	return m
}

func (d *data) ScheduleAssignments(start, end time.Time) []ScheduleAssignment {
	tgtMap := d.rulesByTarget()

	result := make([]ScheduleAssignment, 0, len(tgtMap))
	for tgt, rules := range tgtMap {
		result = append(result, ScheduleAssignment{
			Rules:      rules,
			Target:     tgt,
			Shifts:     d.ShiftsForRules(start, end, rules),
			ScheduleID: d.sched.ID,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		iType := result[i].Target.TargetType()
		jType := result[j].Target.TargetType()
		if iType != jType && iType == assignment.TargetTypeUser {
			return true
		}
		return result[i].Target.TargetID() < result[j].Target.TargetID()
	})

	return result
}

func (d *data) ScheduleFinalShifts(start, end time.Time) []Shift {
	tgtMap := d.rulesByTarget()

	var resultShifts []Shift
	for _, rules := range tgtMap {
		shifts := d.ShiftsForRules(start, end, rules)
		resultShifts = append(resultShifts, shifts...)
	}

	return mergeShiftsByTarget(resultShifts)
}

func (d *data) ShiftsForRules(start, end time.Time, rules []rule.Rule) []Shift {
	var shifts []Shift
	for _, r := range rules {
		shifts = append(shifts, d.ShiftsForRule(start, end, r)...)
	}

	return mergeShiftsByTarget(shifts)
}

func (d *data) rotation(id string) *rotation.Rotation {
	for _, r := range d.rots {
		if r.ID == id {
			return &r
		}
	}

	return nil
}

func (d *data) rotationParticipantUserIDs(id string) []string {
	var parts []rotation.Participant
	for _, p := range d.parts {
		if p.RotationID != id {
			continue
		}
		parts = append(parts, p)
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].Position < parts[j].Position })
	userIDs := make([]string, len(parts))
	for i, p := range parts {
		userIDs[i] = p.Target.TargetID()
	}

	return userIDs
}

func (d *data) rotationState(id string) *rotation.State {
	for _, r := range d.rState {
		if r.RotationID == id {
			return &r
		}
	}

	return nil
}

func (d *data) ShiftsForRule(start, end time.Time, rule rule.Rule) []Shift {
	start = start.In(d.sched.TimeZone)
	end = end.In(d.sched.TimeZone)

	rShifts := ruleShifts(start, end, rule)
	if rule.Target.TargetType() == assignment.TargetTypeUser {
		for i := range rShifts {
			rShifts[i].UserID = rule.Target.TargetID()
		}
		return rShifts
	}

	if len(rShifts) == 0 {
		return nil
	}

	rotID := rule.Target.TargetID()
	state := d.rotationState(rotID)
	if state == nil || !end.After(state.ShiftStart) {
		return nil
	}

	orig := rShifts
	rShifts = rShifts[:0]
	for _, s := range orig {
		if s.End.Before(state.ShiftStart) {
			continue
		}
		rShifts = append(rShifts, s)
	}
	if len(rShifts) == 0 {
		return nil
	}

	userIDs := d.rotationParticipantUserIDs(rotID)
	if len(userIDs) == 0 {
		return nil
	}
	rot := d.rotation(rotID)
	if rot == nil {
		return nil
	}

	partCount := len(userIDs)
	curUserID := userIDs[state.Position%partCount]
	rotEnd := rot.EndTime(state.ShiftStart)
	nextPart := func() {
		state.Position = (state.Position + 1) % partCount
		state.ShiftStart = rotEnd
		curUserID = userIDs[state.Position]
		rotEnd = rot.EndTime(state.ShiftStart)
	}

	for !rotEnd.After(rShifts[0].Start) {
		nextPart()
	}

	expanded := make([]Shift, 0, len(rShifts))
	for _, shift := range rShifts {
		if shift.End.Before(state.ShiftStart) {
			continue
		}
		for !rotEnd.After(shift.Start) {
			nextPart()
		}
		start := shift.Start
		if start.Before(state.ShiftStart) {
			start = state.ShiftStart
		}
		for rotEnd.Before(shift.End) {
			expanded = append(expanded, Shift{Start: start, End: rotEnd, UserID: curUserID})
			start = rotEnd
			nextPart()
		}
		expanded = append(expanded, Shift{Start: start, End: shift.End, UserID: curUserID})
	}
	return expanded
}

func (c *ShiftCalculator) fetchData(ctx context.Context, schedID string) (*data, error) {
	sched, err := c.SchedStore.FindOne(ctx, schedID)
	if err != nil {
		return nil, errors.Wrap(err, "fetch schedule details")
	}
	rules, err := c.RuleStore.FindAll(ctx, schedID)
	if err != nil {
		return nil, errors.Wrap(err, "fetch schedule rules")
	}
	rots, err := c.RotStore.FindAllRotationsByScheduleID(ctx, schedID)
	if err != nil {
		return nil, errors.Wrap(err, "fetch schedule rotations")
	}
	parts, err := c.RotStore.FindAllParticipantsByScheduleID(ctx, schedID)
	if err != nil {
		return nil, errors.Wrap(err, "fetch schedule rotation participants")
	}
	rState, err := c.RotStore.FindAllStateByScheduleID(ctx, schedID)
	if err != nil {
		return nil, errors.Wrap(err, "fetch schedule rotation state")
	}
	return &data{
		sched:  *sched,
		rules:  rules,
		rots:   rots,
		parts:  parts,
		rState: rState,
	}, nil
}

func (c *ShiftCalculator) ScheduleAssignments(ctx context.Context, start, end time.Time, schedID string) ([]ScheduleAssignment, error) {
	data, err := c.fetchData(ctx, schedID)
	if err != nil {
		return nil, err
	}

	return data.ScheduleAssignments(start, end), nil
}

func (c *ShiftCalculator) ScheduleFinalShifts(ctx context.Context, start, end time.Time, schedID string) ([]Shift, error) {
	data, err := c.fetchData(ctx, schedID)
	if err != nil {
		return nil, err
	}

	return data.ScheduleFinalShifts(start, end), nil
}

// ScheduleFinalShiftsWithOverrides will calculate the final set of on-call shifts for the schedule during the given time frame.
func (c *ShiftCalculator) ScheduleFinalShiftsWithOverrides(ctx context.Context, start, end time.Time, schedID string) ([]Shift, error) {
	data, err := c.fetchData(ctx, schedID)
	if err != nil {
		return nil, err
	}
	data.userOverrides, err = c.OStore.FindAllUserOverrides(ctx, start, end, assignment.ScheduleTarget(schedID))
	if err != nil {
		return nil, err
	}

	return data.ScheduleFinalShiftsWithOverrides(start, end), nil
}

// _rotationShifts get's all rotation shifts with hard alignment to start and end.
func _rotationShifts(start, end time.Time, rot *rotation.Rotation, actShiftPos int, actShiftStart time.Time, partIDs []string) []RotationShift {
	if actShiftStart.After(end) {
		return nil
	}
	var shifts []RotationShift

	cPos, cStart, cEnd := actShiftPos, actShiftStart, rot.EndTime(actShiftStart)
	for {
		if cStart.Before(start) {
			cStart = start
		}
		if cEnd.After(end) {
			cEnd = end
		}

		if cEnd.After(start) {
			shifts = append(shifts, RotationShift{
				Start:  cStart,
				End:    cEnd,
				PartID: partIDs[cPos],
			})
		}
		if cEnd.Equal(end) {
			return shifts
		}

		cStart, cEnd, cPos = cEnd, rot.EndTime(cEnd), (cPos+1)%len(partIDs)
	}
}

func (r *ShiftCalculator) RotationShifts(ctx context.Context, start, end time.Time, rotationID string) ([]RotationShift, error) {
	if end.Before(start) {
		return nil, nil
	}
	ctx = log.WithField(ctx, "RotationID", rotationID)
	state, err := r.RotStore.State(ctx, rotationID)
	if err == rotation.ErrNoState {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "lookup rotation state")
	}
	if state == nil {
		return nil, nil
	}

	rot, err := r.RotStore.FindRotation(ctx, rotationID)
	if err != nil {
		return nil, errors.Wrap(err, "lookup rotation config")
	}

	parts, err := r.RotStore.FindAllParticipants(ctx, rotationID)
	if err != nil {
		return nil, errors.Wrap(err, "lookup rotation participants")
	}

	if len(parts) == 0 {
		return nil, nil
	}

	sort.Slice(parts, func(i, j int) bool { return parts[i].Position < parts[j].Position })
	partIDs := make([]string, len(parts))
	for i, p := range parts {
		partIDs[i] = p.ID
	}

	shifts := _rotationShifts(start, end, rot, state.Position, state.ShiftStart, partIDs)
	shifts = mergeRotationShiftsByID(shifts)
	sort.Slice(shifts, func(i, j int) bool { return shifts[i].Start.Before(shifts[j].Start) })

	return shifts, err
}

func mergeRotationShiftsByID(shifts []RotationShift) []RotationShift {
	sort.Slice(shifts, func(i, j int) bool { return shifts[i].Start.Before(shifts[j].Start) })

	m := make(map[string][]RotationShift)
	for _, s := range shifts {
		m[s.PartID] = append(m[s.PartID], s)
	}

	shifts = shifts[:0]
	for _, sh := range m {
		shifts = append(shifts, mergeRotationShifts(sh)...)
	}

	return shifts
}

func mergeRotationShifts(shifts []RotationShift) []RotationShift {
	if len(shifts) < 2 {
		return shifts
	}

	merged := make([]RotationShift, 0, len(shifts))
	cur := shifts[0]
	for _, s := range shifts[1:] {
		if s.Start.Before(cur.End) || s.Start.Equal(cur.End) {
			if s.End.After(cur.End) {
				cur.End = s.End
			}
			continue
		}

		merged = append(merged, cur)
		cur = s
	}
	merged = append(merged, cur)

	return merged
}

func sortShifts(shifts []Shift) {
	sort.Slice(shifts, func(i, j int) bool {
		if !shifts[i].Start.Equal(shifts[j].Start) {
			return shifts[i].Start.Before(shifts[j].Start)
		}
		if !shifts[i].End.Equal(shifts[j].End) {
			return shifts[i].End.Before(shifts[j].End)
		}

		return shifts[i].UserID < shifts[j].UserID
	})
}

func mergeShiftsByTarget(shifts []Shift) []Shift {
	sortShifts(shifts)
	m := make(map[string][]Shift)
	for _, s := range shifts {
		m[s.UserID] = append(m[s.UserID], s)
	}

	shifts = shifts[:0]
	for _, tgtShifts := range m {
		shifts = append(shifts, mergeShifts(tgtShifts)...)
	}
	sortShifts(shifts)
	return shifts
}

// mergeShifts will merge shifts based on start and end times
// s should already be sorted, and it is assumed that all Assignments are identical
func mergeShifts(shifts []Shift) []Shift {
	if len(shifts) < 2 {
		return shifts
	}

	merged := make([]Shift, 0, len(shifts))
	cur := shifts[0]
	for _, s := range shifts[1:] {
		if s.Start.Before(cur.End) || s.Start.Equal(cur.End) {
			if s.End.After(cur.End) {
				cur.End = s.End
			}
			continue
		}

		merged = append(merged, cur)
		cur = s
	}
	merged = append(merged, cur)

	return merged
}

func ruleShifts(start, end time.Time, rule rule.Rule) []Shift {
	if end.Before(start) {
		return nil
	}
	if rule.AlwaysActive() {
		return []Shift{
			{Start: start, End: end},
		}
	}
	if rule.NeverActive() {
		return nil
	}

	var shifts []Shift

	shiftStart := rule.StartTime(start)
	shiftEnd := rule.EndTime(shiftStart)

	var c int
	// arbitrary limit on the number of returned shifts
	for {
		c++
		if c > 10000 {
			panic("too many shifts")
		}
		shifts = append(shifts, Shift{Start: shiftStart, End: shiftEnd})
		shiftStart = rule.StartTime(shiftEnd)
		if shiftStart.After(end) {
			break
		}
		shiftEnd = rule.EndTime(shiftStart)
	}

	return shifts
}
