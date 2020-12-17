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
		- Alert to User H, Service C (created 3nd) -- Not sent, user H already notified
		- Verify to User F
		- Verify to User A -- Not sent, (will get an alert for Service A)
		- Test to User B
		- Alert to User C, Service A
		- Status to User D
		- Status to User G (created 2nd)
	*/

	messages := []Message{

		// Sent
		{
			Type:   TypeTestNotification,
			UserID: "User C",

			// Sent messages are considered regardless of the Dest.Type
			// as the user has been notified *somehow*. That way if we need
			// to make a choice, a user who has gotten no message of any kind
			// would take priority (all other criteria being equal).
			Dest:   notification.Dest{Type: notification.DestTypeVoice, ID: "Voice C"},
			SentAt: n.Add(-2 * time.Minute),
		},
		{
			Type:   TypeTestNotification,
			UserID: "User H",
			Dest:   notification.Dest{Type: notification.DestTypeSMS, ID: "SMS H"},
			SentAt: n.Add(-30 * time.Second),
		},
		{
			Type:      TypeAlertNotification,
			ServiceID: "Service B",
			Dest:      notification.Dest{Type: notification.DestTypeSlackChannel, ID: "Slack B"},
			SentAt:    n.Add(-30 * time.Second),
		},

		// Pending
		{
			ID:        "0",
			Type:      TypeAlertNotification,
			UserID:    "User A",
			ServiceID: "Service A",
			Dest:      notification.Dest{Type: notification.DestTypeSMS, ID: "SMS A"},
			CreatedAt: n,
		}, {
			ID:        "1",
			Type:      TypeAlertNotification,
			UserID:    "User E",
			ServiceID: "Service B",
			Dest:      notification.Dest{Type: notification.DestTypeSMS, ID: "SMS E"},
			CreatedAt: n.Add(1),
		}, {
			// no ID, this message should not be sent this cycle
			Type:      TypeAlertNotification,
			UserID:    "User H",
			ServiceID: "Service C",
			Dest:      notification.Dest{Type: notification.DestTypeSMS, ID: "SMS H"},
			CreatedAt: n.Add(2),
		}, {
			ID:     "2",
			Type:   TypeVerificationMessage,
			UserID: "User F",
			Dest:   notification.Dest{Type: notification.DestTypeSMS, ID: "SMS F"},
		}, {
			// no ID, this message should not be sent this cycle
			Type:   TypeVerificationMessage,
			UserID: "User A",
			Dest:   notification.Dest{Type: notification.DestTypeSMS, ID: "SMS A"},
		}, {
			ID:     "3",
			Type:   TypeTestNotification,
			UserID: "User B",
			Dest:   notification.Dest{Type: notification.DestTypeSMS, ID: "SMS B"},
		}, {
			ID:        "4",
			Type:      TypeAlertNotification,
			UserID:    "User C",
			ServiceID: "Service A",
			Dest:      notification.Dest{Type: notification.DestTypeSMS, ID: "SMS C"},
		}, {
			ID:        "5",
			Type:      TypeAlertStatusUpdate,
			UserID:    "User D",
			Dest:      notification.Dest{Type: notification.DestTypeSMS, ID: "SMS D"},
			CreatedAt: n,
		}, {
			ID:        "6",
			Type:      TypeAlertStatusUpdate,
			UserID:    "User G",
			Dest:      notification.Dest{Type: notification.DestTypeSMS, ID: "SMS G"},
			CreatedAt: n.Add(1),
		}}

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
	for i, exp := range expected {
		msg := q.NextByType(notification.DestTypeSMS)
		require.NotNilf(t, msg, "message #%d", i)
		assert.Equalf(t, exp, *msg, "message #%d", i)
	}

	// no more expected messages
	msg := q.NextByType(notification.DestTypeSMS)
	assert.Nil(t, msg)

}
