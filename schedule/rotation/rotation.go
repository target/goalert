package rotation

import (
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
	"time"
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

func addHours(t time.Time, n int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+n, t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func addHoursAlwaysInc(t time.Time, n int) time.Time {
	res := addHours(t, n)
	if n < 0 {
		for !res.Before(t) {
			n--
			res = addHours(t, n)
		}
	} else {
		for !res.After(t) {
			n++
			res = addHours(t, n)
		}
	}

	return res
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
		return addHoursAlwaysInc(end, -r.ShiftLength)
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
				cTime = addHoursAlwaysInc(cTime, -r.ShiftLength)
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
			cTime = addHoursAlwaysInc(cTime, r.ShiftLength)
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
