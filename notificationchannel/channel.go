package notificationchannel

import (
	"github.com/target/goalert/validation/validate"

	uuid "github.com/satori/go.uuid"
)

type Channel struct {
	ID    string
	Name  string
	Type  Type
	Value string
}

func (c Channel) Normalize() (*Channel, error) {
	if c.ID == "" {
		c.ID = uuid.NewV4().String()
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
