package oncall

import (
	"time"

	"github.com/target/goalert/assignment"
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
			cur := t.Start().In(loc)
			// loop through rotations
			for cur.Before(t.End()) && limit() {
				userID := rule.Rotation.UserID(cur)
				calc.rot.SetSpan(rule.Rotation.CurrentStart, rule.Rotation.CurrentEnd, userID)
				cur = rule.Rotation.CurrentEnd
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
