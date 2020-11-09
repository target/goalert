package rotation

import (
	"time"

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

func addClockHours(t time.Time, n int) time.Time {
	if n == 0 {
		return t
	}
	next := t.Add(time.Duration(n) * time.Hour)

	_, offset := t.Zone()
	_, nextOffset := next.Zone()
	if offset == nextOffset {
		return next
	}

	diffSec := offset - nextOffset
	n *= 3600 // convert to seconds

	if diffSec == -n {
		return next
	}

	return next.Add(time.Duration(diffSec) * time.Second)
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
		return addClockHours(end, -r.ShiftLength)
	case TypeWeekly:
		r.ShiftLength *= 7
	case TypeDaily:
	default:
		panic("unexpected rotation type")
	}

	return end.AddDate(0, 0, -r.ShiftLength)
}

// EndTime calculates the end of the "shift" that started at (or was active)  at t.
//
// For daily and weekly rotations, end time will be the next handoff time (from start).
func (r Rotation) EndTime(t time.Time) time.Time {
	if r.ShiftLength <= 0 {
		r.ShiftLength = 1
	}
	t = t.Truncate(time.Minute)
	cTime := r.Start.Truncate(time.Minute)

	if r.Type == TypeWeekly {
		r.ShiftLength *= 7
	}

	if cTime.After(t) {
		// reverse search
		last := cTime
		switch r.Type {
		case TypeHourly:
			for cTime.After(t) {
				last = cTime
				cTime = addClockHours(cTime, -r.ShiftLength)
			}
		case TypeWeekly, TypeDaily:
			for cTime.After(t) {
				last = cTime
				// getting next end of shift
				cTime = cTime.AddDate(0, 0, -r.ShiftLength)
			}
		default:
			panic("unexpected rotation type")
		}
		return last
	}

	switch r.Type {
	case TypeHourly:
		for !cTime.After(t) {
			cTime = addClockHours(cTime, r.ShiftLength)
		}
	case TypeWeekly, TypeDaily:
		for !cTime.After(t) {
			// getting end of shift
			cTime = cTime.AddDate(0, 0, r.ShiftLength)
		}
	default:
		panic("unexpected rotation type")
	}

	return cTime
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
