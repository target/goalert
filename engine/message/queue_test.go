package message

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/notification/twilio"
)

func voice(value string) gadb.DestV1 {
	return gadb.NewDestV1(twilio.DestTypeTwilioVoice, twilio.FieldPhoneNumber, value)
}

func sms(value string) gadb.DestV1 {
	return gadb.NewDestV1(twilio.DestTypeTwilioSMS, twilio.FieldPhoneNumber, value)
}

func slackCh(value string) gadb.DestV1 {
	return gadb.NewDestV1(slack.DestTypeSlackChannel, slack.FieldSlackChannelID, value)
}

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
			Dest:   voice("Voice C"),
			SentAt: n.Add(-2 * time.Minute),
		},
		{
			Type:   notification.MessageTypeTest,
			UserID: "User H",
			Dest:   sms("SMS H"),
			SentAt: n.Add(-30 * time.Second),
		},
		{
			Type:      notification.MessageTypeAlert,
			ServiceID: "Service B",
			Dest:      slackCh("Slack B"),
			SentAt:    n.Add(-30 * time.Second),
		},

		// Pending
		{
			ID:        "0",
			Type:      notification.MessageTypeAlert,
			UserID:    "User A",
			ServiceID: "Service A",
			Dest:      sms("SMS A"),
			CreatedAt: n,
		},
		{
			ID:        "1",
			Type:      notification.MessageTypeAlert,
			UserID:    "User E",
			ServiceID: "Service B",
			Dest:      sms("SMS E"),
			CreatedAt: n.Add(1),
		},
		{
			// no ID, this message should not be sent this cycle
			Type:      notification.MessageTypeAlert,
			UserID:    "User H",
			ServiceID: "Service C",
			Dest:      sms("SMS H"),
			CreatedAt: n.Add(2),
		},
		{
			ID:     "2",
			Type:   notification.MessageTypeVerification,
			UserID: "User F",
			Dest:   sms("SMS F"),
		},
		{
			// no ID, this message should not be sent this cycle
			Type:   notification.MessageTypeVerification,
			UserID: "User A",
			Dest:   sms("SMS A"),
		},
		{
			ID:     "3",
			Type:   notification.MessageTypeTest,
			UserID: "User B",
			Dest:   sms("SMS B"),
		},
		{
			ID:        "4",
			Type:      notification.MessageTypeAlert,
			UserID:    "User C",
			ServiceID: "Service A",
			Dest:      sms("SMS C"),
		},

		// ThrottleConfig limits 5 messages to be sent in 15 min for DestTypeSMS
		{
			ID:        "5",
			Type:      notification.MessageTypeAlertStatus,
			UserID:    "User D",
			Dest:      sms("SMS D"),
			CreatedAt: n,
		},
		{
			ID:        "6",
			Type:      notification.MessageTypeAlertStatus,
			UserID:    "User G",
			Dest:      sms("SMS G"),
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
	rules := q.cmThrottle.cfg.Rules(Message{Type: notification.MessageTypeAlert, Dest: sms("")})
	expected = expected[:rules[1].Count]

	for i, exp := range expected {
		msg := q.NextByType(twilio.DestTypeTwilioSMS)
		require.NotNilf(t, msg, "message #%d", i)
		assert.Equalf(t, exp, *msg, "message #%d", i)
	}

	// no more expected messages
	msg := q.NextByType(twilio.DestTypeTwilioSMS)
	assert.Nil(t, msg)
}
