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

			// For daily rotations with day restrictions, we need to pause the rotation
			// on inactive days to ensure fair distribution (otherwise some participants
			// would never appear if their turn always falls on inactive days).
			if !rule.IsAlways() && rotCopy.Type == rotation.TypeDaily {
				// For daily rotation with day restrictions, advance position per active day
				activeStart := rule.StartTime(t.Start().In(loc))
				currentIndex := rotCopy.CurrentIndex

				for activeStart.Before(t.End()) && limit() {
					activeEnd := rule.EndTime(activeStart)
					if activeEnd.After(t.End()) {
						activeEnd = t.End()
					}

					userID := rotCopy.Users[currentIndex%len(rotCopy.Users)]
					calc.rot.SetSpan(activeStart, activeEnd, userID)

					currentIndex++

					nextActiveStart := rule.StartTime(activeEnd)
					if nextActiveStart.Equal(activeEnd) || nextActiveStart.Before(activeEnd) {
						break
					}

					activeStart = nextActiveStart
				}
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
