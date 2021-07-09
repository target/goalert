package notificationchannel

import (
	"github.com/google/uuid"
	"github.com/target/goalert/validation/validate"
)

type Channel struct {
	ID    string
	Name  string
	Type  Type
	Value string
}

func (c Channel) Normalize() (*Channel, error) {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}

	err := validate.Many(
		validate.UUID("ID", c.ID),
		validate.Text("Name", c.Name, 1, 255),
		validate.OneOf("Type", c.Type, TypeSlack),
	)

	switch c.Type {
	case TypeSlack:
		err = validate.Many(err, validate.RequiredText("Value", c.Value, 1, 32))
	}

	return &c, err
}
