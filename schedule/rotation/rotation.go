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
	startClock := timeutil.NewClockFromTime(r.Start)

	if r.Type == TypeWeekly {
		r.ShiftLength *= 7
	}
	switch r.Type {
	case TypeHourly:
		startDay := timeutil.StartOfDay(r.Start)
		tDay := timeutil.StartOfDay(t)

		// number of full hours that have passed
		hours := timeutil.HoursBetween(startDay, tDay)

		// the remainder of the shift length
		rem := hours % r.ShiftLength
		if rem != 0 {
			startClock += timeutil.Clock(time.Duration(rem) * time.Hour)
		}
		startHrs := startClock.Hour()
		if startHrs >= 24 {
			whole := startHrs / 24
			tDay = tDay.AddDate(0, 0, whole)
			startClock -= timeutil.Clock(time.Hour * time.Duration(whole*24))
		}

		res := startClock.FirstOfDay(tDay)
		if res.After(t) {
			return res
		}

		return timeutil.AddClock(res, timeutil.NewClock(r.ShiftLength, 0))
	case TypeDaily, TypeWeekly:

		// get the number of full days that have passed
		startDay := timeutil.StartOfDay(r.Start)
		tDay := timeutil.StartOfDay(t)
		days := timeutil.DaysBetween(startDay, tDay)

		// the remainder of the shift length
		rem := days % r.ShiftLength
		if rem != 0 {
			tDay = tDay.AddDate(0, 0, r.ShiftLength-rem)
		}

		res := startClock.FirstOfDay(tDay)
		// already in the future
		if res.After(t) {
			return res
		}

		// t is the day of the handoff, but it has already occured, jump to next
		return startClock.FirstOfDay(res.AddDate(0, 0, r.ShiftLength))
	default:
		panic("unexpected rotation type")
	}
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
