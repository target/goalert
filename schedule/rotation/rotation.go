package rotation

import (
	"time"

	"github.com/target/goalert/util/timeutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

type Rotation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Type           Type      `json:"type"`
	Start          time.Time `json:"start"`
	ShiftLength    int       `json:"shift_length"`
	isUserFavorite bool
}

func (r Rotation) IsUserFavorite() bool {
	return r.isUserFavorite
}

func (r Rotation) shiftClock() timeutil.Clock {
	switch r.Type {
	case TypeHourly:
		return timeutil.NewClock(r.ShiftLength, 0)
	case TypeDaily:
		return timeutil.NewClock(r.ShiftLength*24, 0)
	case TypeWeekly:
		return timeutil.NewClock(r.ShiftLength*24*7, 0)
	default:
		// monthly is handled separately
		panic("unexpected rotation type")
	}
}

func (r Rotation) monthStartTime(t time.Time, n int) time.Time {
	if n > 10000 {
		panic("too many iterations")
	}

	if t.After(r.Start) || t.Equal(r.Start) { // t is after start of rotation
		next := r.Start.AddDate(0, r.ShiftLength*n, 0)
		if next.After(t) {
			return r.Start.AddDate(0, r.ShiftLength*(n-1), 0)
		}

		return r.monthStartTime(t, n+1)
	}

	// t is before start of rotation
	prev := r.Start.AddDate(0, -r.ShiftLength*n, 0)
	if prev.Before(t) {
		return prev
	}

	return r.monthStartTime(t, n+1)
}

func (r Rotation) monthEndTime(t time.Time, n int) time.Time {
	if n > 10000 {
		panic("too many iterations")
	}

	if t.After(r.Start) || t.Equal(r.Start) { // t is after start of rotation
		next := r.Start.AddDate(0, r.ShiftLength*n, 0)
		if next.After(t) {
			return next
		}

		return r.monthEndTime(t, n+1)
	}

	// t is before start of rotation
	prev := r.Start.AddDate(0, -r.ShiftLength*n, 0)
	if prev.Before(t) || prev.Equal(t) {
		return r.Start.AddDate(0, -r.ShiftLength*(n-1), 0)
	}

	return r.monthEndTime(t, n+1)
}

// StartTime calculates the start of the "shift" that started at (or was active) at t.
// For daily and weekly rotations, start time will be the previous handoff time (from start).
func (r Rotation) StartTime(t time.Time) time.Time {
	if r.ShiftLength <= 0 {
		r.ShiftLength = 1
	}
	t = t.In(r.Start.Location()).Truncate(time.Minute)
	r.Start = r.Start.Truncate(time.Minute)

	if r.Type == TypeMonthly {
		return r.monthStartTime(t, 1)
	}

	shiftClockLen := r.shiftClock()
	rem := timeutil.ClockDiff(r.Start, t) % shiftClockLen

	if rem < 0 {
		rem += shiftClockLen
	}

	return timeutil.AddClock(t, -rem)
}

// EndTime calculates the end of the "shift" that started at (or was active) at t.
//
// It is guaranteed to occur after t.
func (r Rotation) EndTime(t time.Time) time.Time {
	if r.ShiftLength <= 0 {
		r.ShiftLength = 1
	}
	t = t.In(r.Start.Location()).Truncate(time.Minute)
	r.Start = r.Start.Truncate(time.Minute)

	if r.Type == TypeMonthly {
		return r.monthEndTime(t, 1)
	}

	shiftClockLen := r.shiftClock()
	rem := timeutil.ClockDiff(r.Start, t) % shiftClockLen

	if rem < 0 {
		rem += shiftClockLen
	}

	return timeutil.AddClock(t, shiftClockLen-rem)
}

func (r Rotation) Normalize() (*Rotation, error) {
	if r.ShiftLength == 0 {
		// default to 1
		r.ShiftLength = 1
	}
	r.Start = r.Start.Truncate(time.Minute)

	if r.Start.Location() == nil {
		return nil, validation.NewFieldError("TimeZone", "must be specified")
	}
	err := validate.Many(
		validate.IDName("Name", r.Name),
		validate.Range("ShiftLength", r.ShiftLength, 1, 9000),
		validate.OneOf("Type", r.Type, TypeMonthly, TypeWeekly, TypeDaily, TypeHourly),
		validate.Text("Description", r.Description, 1, 255),
	)
	if err != nil {
		return nil, err
	}

	return &r, nil
}
