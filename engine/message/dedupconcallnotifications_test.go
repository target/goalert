package message

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
)

func TestDedupOnCallNotifications(t *testing.T) {
	n := time.Now()
	msg := []Message{
		{
			ID:        "a",
			Type:      notification.MessageTypeAlertStatus,
			CreatedAt: n,
		},
		{
			ID:         "b",
			Type:       notification.MessageTypeScheduleOnCallUsers,
			ScheduleID: "A",
			CreatedAt:  n.Add(time.Minute),
		},
		{
			ID:         "c",
			Type:       notification.MessageTypeScheduleOnCallUsers,
			ScheduleID: "A",
			CreatedAt:  n.Add(2 * time.Second),
		},
		{
			ID:         "d",
			Type:       notification.MessageTypeScheduleOnCallUsers,
			ScheduleID: "B",
			CreatedAt:  n.Add(time.Second),
		},
	}

	out, toDelete := dedupOnCallNotifications(msg)
	assert.Equal(t, []string{"c"}, toDelete)

	var ids []string
	for _, m := range out {
		ids = append(ids, m.ID)
	}
	assert.Equal(t, []string{"b", "d", "a"}, ids)
}
