package oncall

import (
	"time"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/schedule/rotation"
)

// SingleRuleCalculator will calculate the currently active user.
type SingleRuleCalculator struct {
	*TimeIterator

	act     *ActiveCalculator
	rot     *UserCalculator
	loc     *time.Location
	rule    ResolvedRule
	userID  string
	changed bool
}

// NewSingleRuleCalculator will create a new SingleRuleCalculator bound to the TimeIterator.
func (t *TimeIterator) NewSingleRuleCalculator(loc *time.Location, rule ResolvedRule) *SingleRuleCalculator {
	calc := &SingleRuleCalculator{
		TimeIterator: t,
		rule:         rule,
		loc:          loc,
		act:          t.NewActiveCalculator(),
	}

	loopLimit := 100000
	limit := func() bool {
		if loopLimit <= 0 {
			panic("infinite loop")
		}
		loopLimit--
		return true
	}

	if rule.AlwaysActive() {
		// always active so just add one span for the entire duration +1 step
		calc.act.SetSpan(t.Start(), t.End().Add(t.Step()))
	} else if !rule.NeverActive() {
		cur := rule.StartTime(t.Start().In(loc))
		// loop through rule active times
		for cur.Before(t.End()) && limit() {
			end := rule.EndTime(cur)
			calc.act.SetSpan(cur, end)
			cur = rule.StartTime(end)
		}
	}
	calc.act.Init()

	if rule.Rotation != nil {
		calc.rot = t.NewUserCalculator()
		switch len(rule.Rotation.Users) {
		case 0:
			// nothing, no on-call
		case 1:
			// always same user
			calc.rot.SetSpan(t.Start(), t.End().Add(t.Step()), rule.Rotation.UserID(t.Start()))
		default:
			rotCopy := *rule.Rotation
			rotCopy.Users = append([]string{}, rule.Rotation.Users...)

			// for daily rotations with day restrictions, we need to pause the rotation
			// on inactive days to ensure fair distribution (otherwise some participants
			// would never appear if their turn always falls on inactive days).
			if !rule.IsAlways() && rotCopy.Type == rotation.TypeDaily {
				calc.setDailyRotationSpans(rotCopy, rule, loc, limit)
			} else {
				cur := t.Start().In(loc)
				// loop through rotations
				for cur.Before(t.End()) && limit() {
					userID := rotCopy.UserID(cur)
					calc.rot.SetSpan(rotCopy.CurrentStart, rotCopy.CurrentEnd, userID)
					cur = rotCopy.CurrentEnd
				}
			}
		}
		calc.rot.Init()
	}

	t.Register(calc)

	return calc
}

// Process implements the SubIterator.Process method.
func (rCalc *SingleRuleCalculator) Process(int64) int64 {
	var newUserID string
	if rCalc.act.Active() {
		if rCalc.rot != nil {
			usrs := rCalc.rot.ActiveUsers()
			if len(usrs) > 0 {
				// rotation will only ever have 1 active user
				newUserID = usrs[0]
			}
		} else if rCalc.rule.Target.TargetType() == assignment.TargetTypeUser {
			newUserID = rCalc.rule.Target.TargetID()
		}
	}

	rCalc.changed = rCalc.userID != newUserID
	rCalc.userID = newUserID

	return 0
}

// Done implements the SubIterator.Done method.
func (rCalc *SingleRuleCalculator) Done() {}

// ActiveUser returns the currently active UserID or an empty string.
func (rCalc *SingleRuleCalculator) ActiveUser() string { return rCalc.userID }

// Changed will return true if the ActiveUser has changed this tick.
func (rCalc *SingleRuleCalculator) Changed() bool { return rCalc.changed }

// setDailyRotationSpans calculates rotation spans for daily rotations with weekday/time restrictions.
// It counts active days from the rotation origin to ensure consistent user assignment across query windows.
func (calc *SingleRuleCalculator) setDailyRotationSpans(rot ResolvedRotation, rule ResolvedRule, loc *time.Location, limit func() bool) {
	// find the first active day at or after rotation start
	origin := rot.Start.In(loc)
	firstActive := rule.StartTime(origin)

	// process each active period in the query window
	start := rule.StartTime(calc.Start().In(loc))

	for start.Before(calc.End()) && limit() {
		end := rule.EndTime(start)
		if end.After(calc.End()) {
			end = calc.End()
		}

		// count active days from rotation origin to determine the correct user
		days := calc.countActiveDays(firstActive, start, rule, loc, limit)

		// calculate user index based on days elapsed from rotation origin
		idx := days % len(rot.Users)
		userID := rot.Users[idx]

		calc.rot.SetSpan(start, end, userID)

		next := rule.StartTime(end)
		if next.Equal(end) || next.Before(end) {
			break
		}

		start = next
	}
}

// countActiveDays counts the number of active days from start up to (but not including) end.
// Times are normalized to midnight for pure calendar day comparison to ensure consistency across query windows.
func (calc *SingleRuleCalculator) countActiveDays(start, end time.Time, rule ResolvedRule, loc *time.Location, limit func() bool) int {
	n := 0
	cur := start

	// normalize end to midnight for comparison
	endMidnight := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, loc)

	for limit() {
		// normalize cur to midnight for pure calendar day comparison
		curMidnight := time.Date(cur.Year(), cur.Month(), cur.Day(), 0, 0, 0, 0, loc)
		if !curMidnight.Before(endMidnight) {
			break
		}

		next := rule.StartTime(rule.EndTime(cur))
		if next.Equal(cur) || next.Before(cur) {
			break
		}
		n++
		cur = next
	}

	return n
}
