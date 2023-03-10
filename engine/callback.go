package engine

import (
	"time"

	"github.com/target/goalert/validation/validate"

	"github.com/google/uuid"
)

type callback struct {
	ID              string
	AlertID         int
	ServiceID       string
	ContactMethodID string
	CreatedAt       time.Time
}

func (c callback) Normalize() (*callback, error) {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	err := validate.Many(
		validate.UUID("ID", c.ID),
		validate.UUID("ContactMethodID", c.ContactMethodID),
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
