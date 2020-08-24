package oncall

import "time"

type RulesCalculator struct {
	*TimeIterator

	rules   []*SingleRuleCalculator
	userIDs []string
	changed bool
}

func (t *TimeIterator) NewRulesCalculator(loc *time.Location, rules []ResolvedRule) *RulesCalculator {
	calc := &RulesCalculator{
		TimeIterator: t,
	}

	for _, r := range rules {
		calc.rules = append(calc.rules, t.NewSingleRuleCalculator(loc, r))
	}
	t.Register(calc.next, nil)

	return calc
}

func (rCalc *RulesCalculator) next(int64) int64 {
	rCalc.changed = false

	for _, r := range rCalc.rules {
		if !r.Changed() {
			continue
		}
		rCalc.changed = true
		break
	}
	if !rCalc.changed {
		return 0
	}

	rCalc.userIDs = rCalc.userIDs[:0]
	for _, r := range rCalc.rules {
		id := r.ActiveUser()
		if id == "" {
			continue
		}
		rCalc.userIDs = append(rCalc.userIDs, id)
	}

	return 0
}

func (rCalc *RulesCalculator) ActiveUsers() []string { return rCalc.userIDs }
func (rCalc *RulesCalculator) Changed() bool         { return rCalc.changed }
