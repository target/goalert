package oncall

import "time"

type SingleRuleCalculator struct {
	*TimeIterator

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
	}
	t.OnNext(calc.next)

	return calc
}

func (rCalc *SingleRuleCalculator) next() {
	newUserID := rCalc.rule.UserID(time.Unix(rCalc.Unix(), 0).In(rCalc.loc))

	rCalc.changed = rCalc.userID != newUserID
	rCalc.userID = newUserID
}

func (rCalc *SingleRuleCalculator) ActiveUser() string { return rCalc.userID }
func (rCalc *SingleRuleCalculator) Changed() bool      { return rCalc.changed }
