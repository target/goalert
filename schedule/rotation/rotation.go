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
		panic("unexpected rotation type")
	}
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
		return timeutil.MonthBeginning(t, *r.Start.Location())
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
		return timeutil.AddClock(timeutil.MonthEnd(t, *r.Start.Location()), timeutil.NewClock(0, 1))
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
