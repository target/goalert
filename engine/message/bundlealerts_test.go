package message

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
)

func TestBundleAlertMessages(t *testing.T) {
	n := time.Now()
	msg := []Message{
		{
			ID:        "a",
			AlertID:   1,
			Type:      notification.MessageTypeAlert,
			CreatedAt: n,
		},
		{
			ID:        "b",
			AlertID:   2,
			Type:      notification.MessageTypeAlert,
			CreatedAt: n.Add(time.Minute),
		},
		{
			ID:        "c",
			AlertID:   3,
			Type:      notification.MessageTypeAlert,
			CreatedAt: n.Add(-time.Hour),
		},
		{
			ID:        "d",
			AlertID:   4,
			Type:      notification.MessageTypeAlert,
			CreatedAt: n.Add(time.Hour),
		},
		{
			ID: "e",
			// bundles for alerts should also be joined
			Type:      notification.MessageTypeAlertBundle,
			CreatedAt: n.Add(time.Hour),
		},
	}

	var bundleID string
	out, err := bundleAlertMessages(msg, func(b Message, ids []string) error {
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
		ID:        bundleID,
		Type:      notification.MessageTypeAlertBundle,
		CreatedAt: n.Add(-time.Hour), // oldest CreatedAt
	}}, out)
}
