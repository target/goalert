package message

import (
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
)

func destIDFrom(s string) notification.DestID {
	return notification.DestID{CMID: uuid.NullUUID{Valid: true, UUID: uuid.Must(uuid.FromBytes([]byte(s)))}}
}

func TestDedupAlerts(t *testing.T) {
	foo := destIDFrom("foofoofoofoofoof")
	bar := destIDFrom("barbarbarbarbarb")
	messages := []Message{
		{ID: "1", Type: notification.MessageTypeTest},
		{ID: "2", Type: notification.MessageTypeAlertBundle},
		{ID: "3", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 11, 0, 0, 0, time.UTC), AlertID: 1, DestID: foo}, // duplicates 4 (same dest and alert, but newer)
		{ID: "4", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 10, 0, 0, 0, time.UTC), AlertID: 1, DestID: foo},
		{ID: "5", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 12, 0, 0, 0, time.UTC), AlertID: 1, DestID: foo}, // duplicates 4 (same dest and alert, but newer)

		{ID: "6", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 8, 0, 0, 0, time.UTC), AlertID: 1, DestID: bar, SentAt: time.Unix(1, 0)}, // duplicates 7 but sent already

		{ID: "7", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 9, 0, 0, 0, time.UTC), AlertID: 1, DestID: bar},
		{ID: "8", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 13, 0, 0, 0, time.UTC), AlertID: 2, DestID: foo},
	}

	res, err := dedupAlerts(messages, func(parentID string, duplicates []string) error {
		assert.Equal(t, "4", parentID)
		sort.Strings(duplicates)
		assert.EqualValues(t, []string{"3", "5"}, duplicates)
		return nil
	})
	assert.NoError(t, err)
	assert.Len(t, res, 6)

	assert.ElementsMatch(t,
		[]Message{
			{ID: "1", Type: notification.MessageTypeTest},
			{ID: "2", Type: notification.MessageTypeAlertBundle},
			{ID: "4", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 10, 0, 0, 0, time.UTC), AlertID: 1, DestID: foo},
			{ID: "7", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 9, 0, 0, 0, time.UTC), AlertID: 1, DestID: bar},
			{ID: "8", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 13, 0, 0, 0, time.UTC), AlertID: 2, DestID: foo},
			{ID: "6", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 8, 0, 0, 0, time.UTC), AlertID: 1, DestID: bar, SentAt: time.Unix(1, 0)},
		},
		res)
}
