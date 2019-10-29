package message

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDB_BundleStatusMessages(t *testing.T) {
	n := time.Now()
	msg := []Message{
		{
			ID:         "a",
			AlertLogID: 5,
			AlertID:    1,
			Type:       TypeAlertStatusUpdate,
			UserID:     "User A",
			CreatedAt:  n,
		},
		{
			ID:         "b",
			AlertLogID: 7,
			AlertID:    2,
			Type:       TypeAlertStatusUpdate,
			UserID:     "User A",
			CreatedAt:  n.Add(time.Minute),
		},
		{
			ID:         "c",
			AlertLogID: 6,
			AlertID:    4,
			Type:       TypeAlertStatusUpdate,
			UserID:     "User A",
			CreatedAt:  n.Add(-time.Hour),
		},
		{
			ID:         "d",
			AlertLogID: 4,
			AlertID:    4,
			Type:       TypeAlertStatusUpdate,
			UserID:     "User A",
			CreatedAt:  n.Add(time.Hour),
		},
		{
			ID:             "e",
			AlertLogID:     3,
			Type:           TypeAlertStatusUpdateBundle,
			UserID:         "User A",
			StatusAlertIDs: []int{7, 8},
			CreatedAt:      n.Add(time.Hour),
		},
	}

	var bundleID string
	out, err := bundleStatusMessages(msg, func(b Message, ids []string) error {
		if bundleID != "" {
			t.Error("got multiple bundles; expected 1")
		}
		bundleID = b.ID
		assert.ElementsMatch(t, []string{"a", "b", "c", "d", "e"}, ids)

		return nil
	})
	assert.NoError(t, err)
	assert.Len(t, out, 1)
	assert.NotEmpty(t, bundleID, "bundled output")
	assert.Equal(t, []Message{{
		ID:             bundleID,
		Type:           TypeAlertStatusUpdateBundle,
		CreatedAt:      n.Add(-time.Hour), // oldest CreatedAt
		AlertLogID:     7,                 // highest ID
		AlertID:        2,                 // Should match Log ID
		UserID:         "User A",
		StatusAlertIDs: []int{1, 2, 4, 7, 8}, // unique alert ids
	}}, out)
}
