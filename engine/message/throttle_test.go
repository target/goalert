package message

import (
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/notification"
	"testing"
	"time"
)

func TestThrottle_Record(t *testing.T) {
	n := time.Now()
	trTime := time.Minute

	var cfg = ThrottleConfig{
		notification.DestTypeSMS: {
			{Count: 1, Per: trTime},
		},
	}

	throttle := NewThrottle(cfg, n, false)

	firstMsg := Message {
		ID:        "0",
		Type:      TypeAlertNotification,
		UserID:    "User A",
		ServiceID: "Service A",
		Dest:      notification.Dest{Type: notification.DestTypeSMS, ID: "SMS A"},
		SentAt: n,
	}

	throttle.Record(firstMsg)
	require.NotNil(t, throttle)
	require.Equal(t, throttle, &Throttle{
		cfg:	cfg,
		ignoreID:	false,
		now:	n,
		count:	map[ThrottleItem]int {
			ThrottleItem{Dest: firstMsg.Dest, BucketDur: trTime}: 1,
		},
		cooldown: map[notification.Dest]bool {
			firstMsg.Dest: true,
		},
	})

	secondMsg := Message {
		ID:        "1",
		Type:      TypeAlertNotification,
		UserID:    "User B",
		ServiceID: "Service B",
		Dest:      notification.Dest{Type: notification.DestTypeSMS, ID: "SMS B"},
		SentAt: n.Add(15 * time.Minute),
	}

	throttle.Record(secondMsg)
	require.NotNil(t, throttle)
	require.Equal(t, throttle, &Throttle{
		cfg:	cfg,
		ignoreID:	false,
		now:	n,
		count:	map[ThrottleItem]int {
			ThrottleItem{Dest: firstMsg.Dest, BucketDur: trTime}: 1,
			ThrottleItem{Dest: secondMsg.Dest, BucketDur: trTime}: 1,
		},
		cooldown: map[notification.Dest]bool {
			firstMsg.Dest: true,
			secondMsg.Dest: true,
		},
	})
}

func TestThrottle_InCooldown(t *testing.T) {
	n := time.Now()

	var cfg = ThrottleConfig{
		notification.DestTypeSMS: {
			{Count: 1, Per: time.Minute},
		},
	}


	throttle := NewThrottle(cfg, n, false)

	msg := Message {
		ID:        "0",
		Type:      TypeAlertNotification,
		UserID:    "User A",
		ServiceID: "Service A",
		Dest:      notification.Dest{Type: notification.DestTypeSMS, ID: "SMS A"},
		SentAt: n,
	}

	pending := throttle.InCooldown(msg)
	require.Equal(t, pending, false)

	throttle.Record(msg)

	pending = throttle.InCooldown(msg)
	require.Equal(t, pending, true)
}
