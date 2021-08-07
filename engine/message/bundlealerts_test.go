package message

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
)

func TestBundleAlertMessages(t *testing.T) {
	t.Run("existing bundle", func(t *testing.T) {

		n := time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)
		bundle := Message{
			ID: "e",
			// bundles for alerts should also be joined
			Type:      notification.MessageTypeAlertBundle,
			CreatedAt: n.Add(time.Hour),
		}
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
			bundle,
		}

		out, err := bundleAlertMessages(msg, func(b Message) (string, error) {
			t.Helper()
			// should use existing bundle
			t.Fail()
			return "", nil
		}, func(parentID string, ids []string) error {
			t.Helper()
			assert.Equal(t, "e", parentID)
			assert.ElementsMatch(t, []string{"a", "b", "c", "d"}, ids)

			return nil
		})
		assert.NoError(t, err)
		assert.Len(t, out, 1)
		assert.EqualValues(t, bundle, out[0])
	})

	t.Run("new bundle", func(t *testing.T) {

		n := time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)

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
		}

		out, err := bundleAlertMessages(msg, func(b Message) (string, error) {
			t.Helper()

			return "e", nil
		}, func(parentID string, ids []string) error {
			t.Helper()
			assert.Equal(t, "e", parentID)
			assert.ElementsMatch(t, []string{"a", "b", "c", "d"}, ids)

			return nil
		})
		assert.NoError(t, err)
		assert.Len(t, out, 1)
		assert.EqualValues(t, Message{
			ID:        "e",
			Type:      notification.MessageTypeAlertBundle,
			CreatedAt: n.Add(-time.Hour),
		}, out[0])
	})

}
