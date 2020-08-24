package oncall

import (
	"strings"

	"github.com/target/goalert/override"
)

type OverrideCalculator struct {
	*TimeIterator

	add     *UserCalculator
	remove  *UserCalculator
	replace *UserCalculator

	userMap map[string]string
}

func (t *TimeIterator) NewOverrideCalculator(overrides []override.UserOverride) *OverrideCalculator {
	calc := &OverrideCalculator{
		TimeIterator: t,
		add:          t.NewUserCalculator(),
		remove:       t.NewUserCalculator(),
		replace:      t.NewUserCalculator(),

		userMap: make(map[string]string),
	}

	for _, o := range overrides {
		if o.AddUserID != "" && o.RemoveUserID != "" {
			calc.replace.SetSpan(o.Start, o.End, o.RemoveUserID+"\n"+o.AddUserID)
		} else if o.AddUserID != "" {
			calc.add.SetSpan(o.Start, o.End, o.AddUserID)
		} else if o.RemoveUserID != "" {
			calc.remove.SetSpan(o.Start, o.End, o.RemoveUserID)
		}
	}
	calc.add.Init()
	calc.remove.Init()
	calc.replace.Init()
	t.Register(calc.next, nil)

	return calc
}
func (oCalc *OverrideCalculator) next() {
	if !oCalc.remove.Changed() && !oCalc.replace.Changed() {
		return
	}

	for id := range oCalc.userMap {
		delete(oCalc.userMap, id)
	}

	for _, id := range oCalc.remove.ActiveUsers() {
		oCalc.userMap[id] = ""
	}
	for _, id := range oCalc.replace.ActiveUsers() {
		parts := strings.SplitN(id, "\n", 2)
		oCalc.userMap[parts[0]] = parts[1]
	}
}

func (oCalc *OverrideCalculator) MapUsers(userIDs []string) []string {
	result := make([]string, 0, len(userIDs))
	for _, id := range userIDs {
		newID, ok := oCalc.userMap[id]
		if !ok {
			result = append(result, id)
			continue
		}
		if newID == "" {
			continue
		}
		result = append(result, newID)
	}

	return append(result, oCalc.add.ActiveUsers()...)
}
