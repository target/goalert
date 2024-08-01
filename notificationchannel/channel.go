package notificationchannel

import (
	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
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
