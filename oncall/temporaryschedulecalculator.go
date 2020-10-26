package oncall

import (
	"github.com/target/goalert/schedule"
)

// TemporaryScheduleCalculator will calculate active state and active users for a set of TemporarySchedules.
type TemporaryScheduleCalculator struct {
	*TimeIterator

	act *ActiveCalculator
	usr *UserCalculator
}

// NewTemporaryScheduleCalculator will create a new TemporaryScheduleCalculator bound to the TimeIterator.
func (t *TimeIterator) NewTemporaryScheduleCalculator(tempScheds []schedule.TemporarySchedule) *TemporaryScheduleCalculator {
	ts := &TemporaryScheduleCalculator{
		TimeIterator: t,
		act:          t.NewActiveCalculator(),
		usr:          t.NewUserCalculator(),
	}

	for _, temp := range tempScheds {
		ts.act.SetSpan(temp.Start, temp.End)

		for _, s := range temp.Shifts {
			ts.usr.SetSpan(s.Start, s.End, s.UserID)
		}
	}
	ts.act.Init()
	ts.usr.Init()

	return ts
}

// Active will return true if a TemporarySchedule is currently active.
func (fg *TemporaryScheduleCalculator) Active() bool { return fg.act.Active() }

// ActiveUsers will return the current set of ActiveUsers. It is only valid if `Active()` is true.
//
// It is only valid if `Active()` is true and until the following Next() call. It should not be modified.
func (fg *TemporaryScheduleCalculator) ActiveUsers() []string { return fg.usr.ActiveUsers() }
