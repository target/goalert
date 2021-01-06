package message_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/engine/message"
	"github.com/target/goalert/notification"
)

func TestThrottle(t *testing.T) {
	n := time.Now()

	cfg := message.ThrottleRules{{Count: 1, Per: time.Minute}}

	throttle := message.NewThrottle(cfg, n, false)

	msg := message.Message{
		ID:        "0",
		Type:      notification.MessageTypeAlert,
		UserID:    "User A",
		ServiceID: "Service A",
		Dest:      notification.Dest{Type: notification.DestTypeSMS, ID: "SMS A"},
		SentAt:    n,
	}

	pending := throttle.InCooldown(msg)
	require.Equal(t, pending, false)

	throttle.Record(msg)

	pending = throttle.InCooldown(msg)
	require.Equal(t, true, pending)
}
