package schedulemanager

import (
	"time"

	"github.com/target/goalert/schedule/rule"
)

// nextUpdateTime returns the next _possible_ update time for the schedule based on its rules, notifications, and overrides.
//
// If there are no upcoming updates, it returns a week from _now_.
func (info *updateInfo) nextUpdateTime(now time.Time) time.Time {
	now = now.In(info.TimeZone)
	next := now.AddDate(0, 0, 7)

	checkTime := func(t time.Time) {
		if t.After(now) && t.Before(next) {
			next = t
		}
	}

	// look for upcoming temp schedule changes
	for _, s := range info.ScheduleData.V1.TemporarySchedules {
		checkTime(s.Start)
		checkTime(s.End)
		for _, r := range s.Shifts {
			checkTime(r.Start)
			checkTime(r.End)
		}
	}

	// look for upcoming overrides
	for _, ovr := range info.Overrides {
		checkTime(ovr.StartTime)
		checkTime(ovr.EndTime)
	}

	// look for upcoming rule changes
	for _, r := range info.Rules {
		rr := rule.RuleFromGADB(r.ScheduleRule)
		// TODO: optimize by using the weekday filter to skip days
		checkTime(rr.StartTime(now))
		checkTime(rr.EndTime(now))
	}

	// look for upcoming notification times
	for _, r := range info.ScheduleData.V1.OnCallNotificationRules {
		if r.NextNotification != nil {
			checkTime(*r.NextNotification)
		}
		if r.Time != nil {
			// TODO: optimize by using the weekday filter to skip days
			checkTime(r.Time.FirstOfDay(now))
			checkTime(r.Time.LastOfDay(now))
		}
	}

	return next
}
