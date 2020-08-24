package oncall

import "time"

type SingleRuleCalculator struct {
	*TimeIterator

	act     *ActiveCalculator
	rot     *UserCalculator
	loc     *time.Location
	rule    ResolvedRule
	userID  string
	changed bool
}

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

	t.Register(calc.next, nil)

	return calc
}

func (rCalc *SingleRuleCalculator) next(int64) {
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
}

func (rCalc *SingleRuleCalculator) ActiveUser() string { return rCalc.userID }
func (rCalc *SingleRuleCalculator) Changed() bool      { return rCalc.changed }
