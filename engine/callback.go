package engine

import (
	"github.com/target/goalert/validation/validate"

	uuid "github.com/satori/go.uuid"
)

type callback struct {
	ID              string
	AlertID         int
	ServiceID       string
	ContactMethodID string
}

func (c callback) Normalize() (*callback, error) {
	if c.ID == "" {
		c.ID = uuid.NewV4().String()
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
