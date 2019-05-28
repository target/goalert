package engine

import (
	"github.com/target/goalert/validation/validate"

	uuid "github.com/satori/go.uuid"
)

type callback struct {
	ID              string
	AlertID         int
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

func (c *callback) fields() []interface{} {
	return []interface{}{
		&c.ID,
		&c.AlertID,
		&c.ContactMethodID,
	}
}
