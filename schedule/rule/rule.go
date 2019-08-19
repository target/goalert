package rule

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/validation/validate"
)

type Rule struct {
	ID         string `json:"id"`
	ScheduleID string `json:"schedule_id"`
	WeekdayFilter
	Start     Clock     `json:"start"`
	End       Clock     `json:"end"`
	CreatedAt time.Time `json:"created_at"`
	Target    assignment.Target
}

func NewAlwaysActive(scheduleID string, tgt assignment.Target) *Rule {
	return &Rule{
		WeekdayFilter: everyDay,
		ScheduleID:    scheduleID,
		Target:        tgt,
	}
}

func (r Rule) Normalize() (*Rule, error) {
	err := validate.UUID("ScheduleID", r.ScheduleID)
	if err != nil {
		return nil, err
	}
	r.Start = Clock(time.Duration(r.Start).Truncate(time.Minute))
	r.End = Clock(time.Duration(r.End).Truncate(time.Minute))
	return &r, nil
}

type scanner interface {
	Scan(...interface{}) error
}

var errNoKnownTarget = errors.New("rule had no known target set (user or rotation)")

func (r *Rule) scanFrom(s scanner) error {
	f := []interface{}{
		&r.ID,
		&r.ScheduleID,
		&r.WeekdayFilter,
		&r.Start,
		&r.End,
	}
	var usr, rot sql.NullString
	f = append(f, &usr, &rot)
	err := s.Scan(f...)
	if err != nil {
		return err
	}

	switch {
	case usr.Valid:
		r.Target = assignment.UserTarget(usr.String)
	case rot.Valid:
		r.Target = assignment.RotationTarget(rot.String)
	default:
		return errNoKnownTarget
	}
	return nil
}
func (r Rule) readFields() []interface{} {
	f := []interface{}{
		&r.ID,
		&r.ScheduleID,
		&r.WeekdayFilter,
		&r.Start,
		&r.End,
	}
	var usr, rot sql.NullString
	switch r.Target.TargetType() {
	case assignment.TargetTypeUser:
		usr.Valid = true
		usr.String = r.Target.TargetID()
	case assignment.TargetTypeRotation:
		rot.Valid = true
		rot.String = r.Target.TargetID()
	}
	return append(f, usr, rot)
}

func (r Rule) everyDay() bool {
	return r.WeekdayFilter == everyDay
}

// StartTime will return the next time the rule would be active.
// If the rule is currently active, it will return the time it
// became active (in the past).
//
// If the rule is NeverActive or AlwaysActive, zero time is returned.
func (r Rule) StartTime(t time.Time) time.Time {
	if r.NeverActive() {
		return time.Time{}
	}
	if r.AlwaysActive() {
		return time.Time{}
	}
	t = t.Truncate(time.Minute)
	start := time.Date(t.Year(), t.Month(), t.Day(), r.Start.Hour(), r.Start.Minute(), 0, 0, t.Location())

	if r.IsActive(t) {
		if start.After(t) {
			start = start.AddDate(0, 0, -1)
		}
		if r.everyDay() {
			return start
		}
		if r.Start == r.End {
			start = start.AddDate(0, 0, -r.DaysSince(start.Weekday(), false)+1)
		}
	} else {
		if start.Before(t) {
			start = start.AddDate(0, 0, 1)
		}
		if r.everyDay() {
			return start
		}

		start = start.AddDate(0, 0, r.DaysUntil(start.Weekday(), true))
	}

	return start
}

// EndTime will return the next time the rule would be inactive.
// If the rule is currently inactive, it will return the end
// of the next shift.
func (r Rule) EndTime(t time.Time) time.Time {
	if r.NeverActive() {
		return time.Time{}
	}
	if r.AlwaysActive() {
		return time.Time{}
	}

	start := r.StartTime(t)
	end := time.Date(start.Year(), start.Month(), start.Day(), r.End.Hour(), r.End.Minute(), 0, 0, t.Location())
	if !end.After(start) {
		end = end.AddDate(0, 0, 1)
	}

	if r.everyDay() {
		return end
	}

	if r.Start == r.End {
		end = end.AddDate(0, 0, r.DaysUntil(start.Weekday(), false)-1)
	}

	return end
}

// NeverActive returns true if the rule will never be active.
func (r Rule) NeverActive() bool { return r.WeekdayFilter == neverDays }

// AlwaysActive will return true if the rule will always be active.
func (r Rule) AlwaysActive() bool { return r.WeekdayFilter == everyDay && r.Start == r.End }

// IsActive determines if the rule is active in the given moment in time, in the location of t.
func (r Rule) IsActive(t time.Time) bool {
	if r.NeverActive() {
		return false
	}
	if r.AlwaysActive() {
		return true
	}
	t = t.Truncate(time.Minute)

	c := NewClock(t.Hour(), t.Minute())
	if r.Start >= r.End { // overnight
		prevDay := (t.Weekday() - 1) % 7
		if prevDay < 0 {
			prevDay += 7
		}
		return (r.Day(t.Weekday()) && c >= r.Start) || (r.Day(prevDay) && c < r.End)
	}

	return r.Day(t.Weekday()) && c >= r.Start && c < r.End
}

// String returns a human-readable string describing the rule
func (r Rule) String() string {
	if r.AlwaysActive() {
		return "Always"
	}
	if r.NeverActive() {
		return "Never"
	}

	var startStr, endStr string
	if r.Start.Minute() == 0 {
		startStr = r.Start.Format("3pm")
	} else {
		startStr = r.Start.Format("3:04pm")
	}

	if r.End.Minute() == 0 {
		endStr = r.End.Format("3pm")
	} else {
		endStr = r.End.Format("3:04pm")
	}

	return fmt.Sprintf("%s-%s %s", startStr, endStr, r.WeekdayFilter.String())
}
