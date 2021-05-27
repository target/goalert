package message

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
)

func TestDB_BundleStatusMessages(t *testing.T) {
	n := time.Now()
	msg := []Message{
		{
			ID:         "a",
			AlertLogID: 5,
			AlertID:    1,
			Type:       notification.MessageTypeAlertStatus,
			UserID:     "User A",
			CreatedAt:  n,
		},
		{
			ID:         "b",
			AlertLogID: 7,
			AlertID:    2,
			Type:       notification.MessageTypeAlertStatus,
			UserID:     "User A",
			CreatedAt:  n.Add(time.Minute),
		},
		{
			ID:         "c",
			AlertLogID: 6,
			AlertID:    4,
			Type:       notification.MessageTypeAlertStatus,
			UserID:     "User A",
			CreatedAt:  n.Add(-time.Hour),
		},
		{
			ID:         "d",
			AlertLogID: 4,
			AlertID:    4,
			Type:       notification.MessageTypeAlertStatus,
			UserID:     "User A",
			CreatedAt:  n.Add(time.Hour),
		},
	}

	out, toDelete := dedupStatusMessages(msg)
	assert.Len(t, out, 3, "output messages")
	assert.Len(t, toDelete, 1, "to delete")
	assert.Equal(t, []string{"d"}, toDelete)

	assert.Equal(t, []Message{
		{
			ID:         "b",
			AlertLogID: 7,
			AlertID:    2,
			Type:       notification.MessageTypeAlertStatus,
			UserID:     "User A",
			CreatedAt:  n.Add(time.Minute),
		},
		{
			ID:         "c",
			AlertLogID: 6,
			AlertID:    4,
			Type:       notification.MessageTypeAlertStatus,
			UserID:     "User A",
			CreatedAt:  n.Add(-time.Hour),
		},
		{
			ID:         "a",
			AlertLogID: 5,
			AlertID:    1,
			Type:       notification.MessageTypeAlertStatus,
			UserID:     "User A",
			CreatedAt:  n,
		},
	}, out)
}
