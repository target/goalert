package schedule

import (
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
	"time"
)

type Schedule struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	TimeZone    *time.Location `json:"time_zone"`
	isUserFavorite bool
}

func (s Schedule) Normalize() (*Schedule, error) {
	err := validate.Many(
		validate.IDName("Name", s.Name),
		validate.Text("Description", s.Description, 1, 255),
	)
	if err != nil {
		return nil, err
	}

	if s.TimeZone == nil {
		return nil, validation.NewFieldError("TimeZone", "must be specified")
	}

	return &s, nil
}
// IsUserFavorite returns a boolean value based on if the schedule is a favorite of the user or not.
func (s Schedule) IsUserFavorite() bool {
	return s.isUserFavorite
}