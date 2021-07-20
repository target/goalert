package message

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
)

func TestDedupAlerts(t *testing.T) {
	messages := []Message{
		{ID: "1", Type: notification.MessageTypeTest},
		{ID: "2", Type: notification.MessageTypeAlertBundle},
		{ID: "3", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 11, 0, 0, 0, time.UTC), AlertID: 1, Dest: notification.Dest{ID: "foo"}}, // duplicates 4 (same dest and alert, but newer)
		{ID: "4", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 10, 0, 0, 0, time.UTC), AlertID: 1, Dest: notification.Dest{ID: "foo"}},
		{ID: "5", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 12, 0, 0, 0, time.UTC), AlertID: 1, Dest: notification.Dest{ID: "foo"}}, // duplicates 4 (same dest and alert, but newer)
		{ID: "6", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 9, 0, 0, 0, time.UTC), AlertID: 1, Dest: notification.Dest{ID: "bar"}},
		{ID: "7", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 13, 0, 0, 0, time.UTC), AlertID: 2, Dest: notification.Dest{ID: "foo"}},
	}

	res, err := dedupAlerts(messages, func(parentID string, duplicates []string) error {
		assert.Equal(t, "4", parentID)
		sort.Strings(duplicates)
		assert.EqualValues(t, []string{"3", "5"}, duplicates)
		return nil
	})
	assert.NoError(t, err)
	assert.Len(t, res, 5)
	sort.Slice(res, func(i, j int) bool { return res[i].ID < res[j].ID })

	assert.EqualValues(t,
		[]Message{
			{ID: "1", Type: notification.MessageTypeTest},
			{ID: "2", Type: notification.MessageTypeAlertBundle},
			{ID: "4", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 10, 0, 0, 0, time.UTC), AlertID: 1, Dest: notification.Dest{ID: "foo"}},
			{ID: "6", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 9, 0, 0, 0, time.UTC), AlertID: 1, Dest: notification.Dest{ID: "bar"}},
			{ID: "7", Type: notification.MessageTypeAlert, CreatedAt: time.Date(2021, 7, 15, 13, 0, 0, 0, time.UTC), AlertID: 2, Dest: notification.Dest{ID: "foo"}},
		},
		res)
}
