package oncall

import "time"

// RulesCalculator provides a single interface for calculating active users of a set of ResolvedRules.
type RulesCalculator struct {
	*TimeIterator

	rules   []*SingleRuleCalculator
	userIDs []string
	changed bool
}

// NewRulesCalculator will create a new RulesCalculator bound to the TimeIterator.
func (t *TimeIterator) NewRulesCalculator(loc *time.Location, rules []ResolvedRule) *RulesCalculator {
	calc := &RulesCalculator{
		TimeIterator: t,
	}

	for _, r := range rules {
		calc.rules = append(calc.rules, t.NewSingleRuleCalculator(loc, r))
	}
	t.Register(calc)

	return calc
}

// Process implements the SubIterator.Process method.
func (rCalc *RulesCalculator) Process(int64) int64 {
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

// Done implements the SubIterator.Done method.
func (rCalc *RulesCalculator) Done() {}

// ActiveUsers returns the current set of active users for the current timestamp.
// It is only valid until the following Next() call and should not be modified.
func (rCalc *RulesCalculator) ActiveUsers() []string { return rCalc.userIDs }

// Changed will return true if there has been any change this tick.
func (rCalc *RulesCalculator) Changed() bool { return rCalc.changed }
