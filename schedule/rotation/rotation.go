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

// StartTime calculates the start of the "shift" that started at (or was active) at t.
// For daily and weekly rotations, start time will be the previous handoff time (from start).
func (r Rotation) StartTime(t time.Time) time.Time {
	if r.ShiftLength <= 0 {
		r.ShiftLength = 1
	}
	end := r.EndTime(t)

	switch r.Type {
	case TypeHourly:
		return timeutil.AddClock(end, timeutil.NewClock(-r.ShiftLength, 0))
	case TypeWeekly:
		r.ShiftLength *= 7
	case TypeDaily:
	default:
		panic("unexpected rotation type")
	}

	end = timeutil.StartOfDay(end).AddDate(0, 0, -r.ShiftLength)
	return timeutil.NewClockFromTime(r.Start).FirstOfDay(end)
}

// EndTime calculates the end of the "shift" that started at (or was active) at t.
//
// It is guaranteed to occur after t.
func (r Rotation) EndTime(t time.Time) time.Time {
	if r.ShiftLength <= 0 {
		r.ShiftLength = 1
	}
	t = t.Truncate(time.Minute)
	r.Start = r.Start.Truncate(time.Minute)

	switch r.Type {
	case TypeHourly:
	case TypeDaily:
		r.ShiftLength *= 24
	case TypeWeekly:
		r.ShiftLength *= 24 * 7
	default:
		panic("unexpected rotation type")
	}

	// number of clock hours that have passed
	hours := timeutil.HoursBetween(r.Start, t)

	// add the remainder of the shift length to get the next shift
	rem := hours % r.ShiftLength
	hours += r.ShiftLength - rem

	return timeutil.AddClock(r.Start, timeutil.NewClock(hours, 0))
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
		validate.OneOf("Type", r.Type, TypeWeekly, TypeDaily, TypeHourly),
		validate.Text("Description", r.Description, 1, 255),
	)
	if err != nil {
		return nil, err
	}

	return &r, nil
}
