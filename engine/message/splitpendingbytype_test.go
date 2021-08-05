package message

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
)

func TestSplitPendingByType(t *testing.T) {
	msgs := []Message{
		{SentAt: time.Unix(1, 0), Type: notification.MessageTypeAlert},
		{Type: notification.MessageTypeAlertBundle},
		{Type: notification.MessageTypeAlert},
		{Type: notification.MessageTypeTest},
	}

	match, remainder := splitPendingByType(msgs, notification.MessageTypeAlertBundle, notification.MessageTypeTest)
	assert.ElementsMatch(t, []Message{
		{Type: notification.MessageTypeAlertBundle},
		{Type: notification.MessageTypeTest},
	}, match)
	assert.ElementsMatch(t, []Message{
		{SentAt: time.Unix(1, 0), Type: notification.MessageTypeAlert},
		{Type: notification.MessageTypeAlert},
	}, remainder)

}
