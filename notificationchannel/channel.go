package notificationchannel

import (
	"strings"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/validation/validate"
)

type Channel struct {
	ID    uuid.UUID
	Name  string
	Type  Type
	Value string
}

func (c *Channel) fromRow(row gadb.NotificationChannel) {
	c.ID = row.ID
	c.Name = row.Name
	c.Type = Type(row.Type)
	c.Value = row.Value
}

func (c Channel) Normalize() (*Channel, error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}

	err := validate.Many(
		validate.Text("Name", c.Name, 1, 255),
		validate.OneOf("Type", c.Type, TypeSlackChan, TypeWebhook, TypeSlackUG),
	)

	switch c.Type {
	case TypeSlackUG:
		grp, ch, _ := strings.Cut(c.Value, ":")
		err = validate.Many(err,
			validate.RequiredText("Value.GroupID", grp, 1, 32),
			validate.RequiredText("Value.ChannelID", ch, 1, 32),
		)
	case TypeSlackChan:
		err = validate.Many(err, validate.RequiredText("Value", c.Value, 1, 32))
	case TypeWebhook:
		err = validate.Many(err, validate.URL("Value", c.Value))
	}

	return &c, err
}
