package message_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/engine/message"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/twilio"
)

func TestThrottleConfigBuilder(t *testing.T) {
	var b message.ThrottleConfigBuilder

	b.AddRules([]message.ThrottleRule{{Count: 1, Per: 2 * time.Minute}})

	b.WithDestTypes(twilio.DestTypeTwilioSMS).AddRules([]message.ThrottleRule{{Count: 2, Per: 3 * time.Minute}})

	b.WithMsgTypes(notification.MessageTypeAlert).AddRules([]message.ThrottleRule{{Count: 3, Per: 5 * time.Minute}})

	b.WithDestTypes(twilio.DestTypeTwilioVoice).WithMsgTypes(notification.MessageTypeTest).AddRules([]message.ThrottleRule{{Count: 5, Per: 7 * time.Minute}})

	cfg := b.Config()

	assert.Equal(t, 7*time.Minute, cfg.MaxDuration())

	check := func(dest string, msg notification.MessageType, expRules []message.ThrottleRule) {
		t.Helper()
		assert.EqualValues(t, expRules, cfg.Rules(message.Message{Type: msg, Dest: gadb.DestV1{Type: dest}}))
	}

	check(
		"",
		notification.MessageTypeUnknown,
		[]message.ThrottleRule{
			{Count: 1, Per: 2 * time.Minute},
		},
	)

	check(
		twilio.DestTypeTwilioSMS,
		notification.MessageTypeUnknown,
		[]message.ThrottleRule{
			{Count: 1, Per: 2 * time.Minute},
			{Count: 2, Per: 3 * time.Minute},
		},
	)

	check(
		twilio.DestTypeTwilioVoice,
		notification.MessageTypeAlert,
		[]message.ThrottleRule{
			{Count: 1, Per: 2 * time.Minute},
			{Count: 3, Per: 5 * time.Minute},
		},
	)

	check(
		twilio.DestTypeTwilioVoice,
		notification.MessageTypeTest,
		[]message.ThrottleRule{
			{Count: 1, Per: 2 * time.Minute},
			{Count: 5, Per: 7 * time.Minute},
		},
	)

	check(
		twilio.DestTypeTwilioSMS,
		notification.MessageTypeAlert,
		[]message.ThrottleRule{
			{Count: 1, Per: 2 * time.Minute},
			{Count: 2, Per: 3 * time.Minute},
			{Count: 3, Per: 5 * time.Minute},
		},
	)
}
