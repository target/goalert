package oncall

import (
	"github.com/target/goalert/schedule"
)

// FixedGroupCalculator will calculate active state and active users for a set of FixedShiftGroups.
type FixedGroupCalculator struct {
	*TimeIterator

	act *ActiveCalculator
	usr *UserCalculator
}

// NewFixedGroupCalculator will create a new FixedGroupCalculator bound to the TimeIterator.
func (t *TimeIterator) NewFixedGroupCalculator(groups []schedule.FixedShiftGroup) *FixedGroupCalculator {
	fg := &FixedGroupCalculator{
		TimeIterator: t,
		act:          t.NewActiveCalculator(),
		usr:          t.NewUserCalculator(),
	}

	for _, g := range groups {
		fg.act.SetSpan(g.Start, g.End)

		for _, s := range g.Shifts {
			fg.usr.SetSpan(s.Start, s.End, s.UserID)
		}
	}
	fg.act.Init()
	fg.usr.Init()

	return fg
}

// Active will return true if a FixedShiftGroup is currently active.
func (fg *FixedGroupCalculator) Active() bool { return fg.act.Active() }

// ActiveUsers will return the current set of ActiveUsers. It is only valid if `Active()` is true.
//
// It is only valid if `Active()` is true and until the following Next() call. It should not be modified.
func (fg *FixedGroupCalculator) ActiveUsers() []string { return fg.usr.ActiveUsers() }
