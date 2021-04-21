package oncall

import (
	"sort"
	"strings"

	"github.com/target/goalert/override"
)

// OverrideCalculator allows mapping a set of active users against the current set of active overrides.
type OverrideCalculator struct {
	*TimeIterator

	add     *UserCalculator
	remove  *UserCalculator
	replace *UserCalculator

	userMap map[string]string

	mapUsers []string
}

// NewOverrideCalculator will create a new OverrideCalculator bound to the TimeIterator.
func (t *TimeIterator) NewOverrideCalculator(overrides []override.UserOverride) *OverrideCalculator {
	calc := &OverrideCalculator{
		TimeIterator: t,
		add:          t.NewUserCalculator(),
		remove:       t.NewUserCalculator(),
		replace:      t.NewUserCalculator(),

		userMap:  make(map[string]string),
		mapUsers: make([]string, 0, 20),
	}

	sort.Slice(overrides, func(i, j int) bool { return overrides[i].Start.Before(overrides[j].Start) })
	for _, o := range overrides {
		if o.AddUserID != "" && o.RemoveUserID != "" {
			// We need both remove & add, so store them with a newline separator REMOVE/REPLACE
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
	t.Register(calc)

	return calc
}

// Process implements the SubIterator.Process method.
func (oCalc *OverrideCalculator) Process(int64) int64 {
	if !oCalc.remove.Changed() && !oCalc.replace.Changed() {
		// add is always used, so if neither of these has changed, there's nothing to do.
		return 0
	}

	// clear the existing map
	for id := range oCalc.userMap {
		delete(oCalc.userMap, id)
	}

	for _, id := range oCalc.remove.ActiveUsers() {
		oCalc.userMap[id] = ""
	}
	for _, id := range oCalc.replace.ActiveUsers() {
		parts := strings.SplitN(id, "\n", 2)
		// REMOVE/REPLACE
		oCalc.userMap[parts[0]] = parts[1]
	}

	return 0
}

// Done implements the SubIterator.Done method.
func (oCalc *OverrideCalculator) Done() {}

// MapUsers will return a new slice of userIDs, taking into account any active overrides.
//
// It is only valid until the following Next() call and should not be modified.
func (oCalc *OverrideCalculator) MapUsers(userIDs []string) []string {
	// re-use existing slice
	oCalc.mapUsers = oCalc.mapUsers[:0]
	for _, id := range userIDs {
		newID, ok := oCalc.userMap[id]
		if !ok {
			oCalc.mapUsers = append(oCalc.mapUsers, id)
			continue
		}
		if newID == "" {
			continue
		}
		oCalc.mapUsers = append(oCalc.mapUsers, newID)
	}

	oCalc.mapUsers = append(oCalc.mapUsers, oCalc.add.ActiveUsers()...)

	return oCalc.mapUsers
}
