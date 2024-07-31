package notificationchannel

import (
	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/validation/validate"
)

type Channel struct {
	ID   uuid.UUID
	Name string
	Dest gadb.DestV1
}

func (c *Channel) fromRow(row gadb.NotificationChannel) {
	c.ID = row.ID
	c.Name = row.Name
	c.Dest = row.Dest.DestV1
}

func (c Channel) Normalize() (*Channel, error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}

	err := validate.Text("Name", c.Name, 1, 255)
	return &c, err
}
