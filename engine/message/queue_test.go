package message

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/notification"
)

func TestQueue_Sort(t *testing.T) {
	n := time.Now()

	/*
		Sent:
		- Test to User C
		- Verify to User H (< 60 seconds ago)
		- Alert to Slack for Service B

		Pending:
		- Alert to User A, Service A (first alert)
		- Alert to User E, Service B (created 2nd -- SMS so Slack doesn't affect first alert status)
		- Alert to User H, Service C (created 3rd) -- Not sent, user H already notified
		- Verify to User F
		- Verify to User A -- Not sent, (will get an alert for Service A)
		- Test to User B
		- Alert to User C, Service A

		Throttled:
		- Status to User D
		- Status to User G (created 2nd)
	*/

	messages := []Message{
		// Sent
		{
			Type:   notification.MessageTypeTest,
			UserID: "User C",

			// Sent messages are considered regardless of the Dest.Type
			// as the user has been notified *somehow*. That way if we need
			// to make a choice, a user who has gotten no message of any kind
			// would take priority (all other criteria being equal).
			Dest:   notification.Dest{Type: notification.DestTypeVoice, Value: "Voice C"},
			SentAt: n.Add(-2 * time.Minute),
		},
		{
			Type:   notification.MessageTypeTest,
			UserID: "User H",
			Dest:   notification.Dest{Type: notification.DestTypeSMS, Value: "SMS H"},
			SentAt: n.Add(-30 * time.Second),
		},
		{
			Type:      notification.MessageTypeAlert,
			ServiceID: "Service B",
			Dest:      notification.Dest{Type: notification.DestTypeSlackChannel, Value: "Slack B"},
			SentAt:    n.Add(-30 * time.Second),
		},

		// Pending
		{
			ID:        "0",
			Type:      notification.MessageTypeAlert,
			UserID:    "User A",
			ServiceID: "Service A",
			Dest:      notification.Dest{Type: notification.DestTypeSMS, Value: "SMS A"},
			CreatedAt: n,
		},
		{
			ID:        "1",
			Type:      notification.MessageTypeAlert,
			UserID:    "User E",
			ServiceID: "Service B",
			Dest:      notification.Dest{Type: notification.DestTypeSMS, Value: "SMS E"},
			CreatedAt: n.Add(1),
		},
		{
			// no ID, this message should not be sent this cycle
			Type:      notification.MessageTypeAlert,
			UserID:    "User H",
			ServiceID: "Service C",
			Dest:      notification.Dest{Type: notification.DestTypeSMS, Value: "SMS H"},
			CreatedAt: n.Add(2),
		},
		{
			ID:     "2",
			Type:   notification.MessageTypeVerification,
			UserID: "User F",
			Dest:   notification.Dest{Type: notification.DestTypeSMS, Value: "SMS F"},
		},
		{
			// no ID, this message should not be sent this cycle
			Type:   notification.MessageTypeVerification,
			UserID: "User A",
			Dest:   notification.Dest{Type: notification.DestTypeSMS, Value: "SMS A"},
		},
		{
			ID:     "3",
			Type:   notification.MessageTypeTest,
			UserID: "User B",
			Dest:   notification.Dest{Type: notification.DestTypeSMS, Value: "SMS B"},
		},
		{
			ID:        "4",
			Type:      notification.MessageTypeAlert,
			UserID:    "User C",
			ServiceID: "Service A",
			Dest:      notification.Dest{Type: notification.DestTypeSMS, Value: "SMS C"},
		},

		// ThrottleConfig limits 5 messages to be sent in 15 min for DestTypeSMS
		{
			ID:        "5",
			Type:      notification.MessageTypeAlertStatus,
			UserID:    "User D",
			Dest:      notification.Dest{Type: notification.DestTypeSMS, Value: "SMS D"},
			CreatedAt: n,
		},
		{
			ID:        "6",
			Type:      notification.MessageTypeAlertStatus,
			UserID:    "User G",
			Dest:      notification.Dest{Type: notification.DestTypeSMS, Value: "SMS G"},
			CreatedAt: n.Add(1),
		},
	}

	var expected []Message
	for _, m := range messages {
		if !m.SentAt.IsZero() || m.ID == "" {
			continue
		}
		if m.ID != strconv.Itoa(len(expected)) {
			t.Fatal("expected messages must be in order (starting with 0)")
		}
		expected = append(expected, m)
	}

	// shuffle order for testing
	rand.Shuffle(len(messages), func(i, j int) { messages[i], messages[j] = messages[j], messages[i] })

	q := newQueue(messages, n)

	// limit the number expected messages to the number allowed to be sent in 15 min
	rules := q.cmThrottle.cfg.Rules(Message{Type: notification.MessageTypeAlert, Dest: notification.Dest{Type: notification.DestTypeSMS}})
	expected = expected[:rules[1].Count]

	for i, exp := range expected {
		msg := q.NextByType(notification.DestTypeSMS.String())
		require.NotNilf(t, msg, "message #%d", i)
		assert.Equalf(t, exp, *msg, "message #%d", i)
	}

	// no more expected messages
	msg := q.NextByType(notification.DestTypeSMS.String())
	assert.Nil(t, msg)
}
