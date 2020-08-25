package oncall

import "time"

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

	if rule.AlwaysActive() {
		calc.act.SetSpan(t.Start(), t.End().Add(t.Step()))
	} else if !rule.NeverActive() {
		cur := rule.StartTime(t.Start().In(loc))
		for cur.Before(t.End()) {
			end := rule.EndTime(cur)
			calc.act.SetSpan(cur, end)
			cur = rule.StartTime(end)
		}
	}
	calc.act.Init()

	if rule.Rotation != nil {
		calc.rot = t.NewUserCalculator()
		cur := t.Start().In(loc)
		for cur.Before(t.End()) {
			userID := rule.Rotation.UserID(cur)
			calc.rot.SetSpan(rule.Rotation.CurrentStart, rule.Rotation.CurrentEnd, userID)
			cur = rule.Rotation.CurrentEnd
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
				newUserID = usrs[0]
			}
		} else {
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
