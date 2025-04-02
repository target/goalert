package rule

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/util/timeutil"
	"github.com/target/goalert/validation/validate"
)

type Rule struct {
	ID         string `json:"id"`
	ScheduleID string `json:"schedule_id"`
	timeutil.WeekdayFilter
	Start     timeutil.Clock `json:"start"`
	End       timeutil.Clock `json:"end"`
	CreatedAt time.Time      `json:"created_at"`
	Target    assignment.Target
}

func NewAlwaysActive(scheduleID string, tgt assignment.Target) *Rule {
	return &Rule{
		WeekdayFilter: timeutil.EveryDay(),
		ScheduleID:    scheduleID,
		Target:        tgt,
	}
}

func (r Rule) Normalize() (*Rule, error) {
	err := validate.UUID("ScheduleID", r.ScheduleID)
	if err != nil {
		return nil, err
	}
	r.Start = timeutil.Clock(time.Duration(r.Start).Truncate(time.Minute))
	r.End = timeutil.Clock(time.Duration(r.End).Truncate(time.Minute))
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

// StartTime will return the next time the rule would be active.
// If the rule is currently active, it will return the time it
// became active (in the past).
//
// If the rule is NeverActive or AlwaysActive, zero time is returned.
//
// It may break when processing a timezone where daylight savings repeats or skips
// ahead at midnight.
func (r Rule) StartTime(t time.Time) time.Time {
	if r.NeverActive() {
		return time.Time{}
	}
	if r.AlwaysActive() {
		return time.Time{}
	}
	t = t.Truncate(time.Minute)

	w := t.Weekday()
	isTodayEnabled := r.Day(w)

	if isTodayEnabled && r.Start == r.End {
		return r.Start.FirstOfDay(r.StartTime(t))
	}

	if r.Start < r.End {
		if !isTodayEnabled {
			return r.Start.FirstOfDay(r.NextActive(t))
		}

		// same-day shift, was active today
		// see if it's already ended
		end := r.End.LastOfDay(t)
		if t.Before(end) {
			return r.Start.FirstOfDay(t)
		}

		// shift has ended for the day, find the next start
		return r.Start.FirstOfDay(r.NextActive(t))
	}

	end := r.End.LastOfDay(t)
	if r.Day(w-1) && t.Before(end) {
		// yesterday's shift is still active
		return r.Start.FirstOfDay(t.AddDate(0, 0, -1))
	}

	if isTodayEnabled {
		// started or will start today
		return r.Start.FirstOfDay(t)
	}

	// find the next start time
	return r.Start.FirstOfDay(r.NextActive(t))
}

// EndTime will return the next time the rule would be inactive.
// If the rule is currently inactive, it will return the end
// of the next shift.
//
// If the rule is always active, or never active, it returns a zero time.
func (r Rule) EndTime(t time.Time) time.Time {
	start := r.StartTime(t)
	if start.IsZero() {
		return start
	}

	if r.Start < r.End {
		return r.End.LastOfDay(start)
	}
	if r.Start > r.End {
		// always the day after the start
		return r.End.LastOfDay(start.AddDate(0, 0, 1))
	}

	// 24-hour rule, end time of the next inactive day
	return r.End.LastOfDay(r.NextInactive(start))
}

// NeverActive returns true if the rule will never be active.
func (r Rule) NeverActive() bool { return r.IsNever() }

// AlwaysActive will return true if the rule will always be active.
func (r Rule) AlwaysActive() bool { return r.IsAlways() && r.Start == r.End }

// IsActive determines if the rule is active in the given moment in time, in the location of t.
func (r Rule) IsActive(t time.Time) bool {
	if r.NeverActive() {
		return false
	}
	if r.AlwaysActive() {
		return true
	}

	return !r.StartTime(t).After(t)
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
