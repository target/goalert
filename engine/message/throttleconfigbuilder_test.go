package message_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/engine/message"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfy"
)

const (
	TestTypeVoice = "test-type-voice"
	TestTypeSMS   = "test-type-sms"
	TestTypeSlack = "test-type-slack"
)

func TestThrottleConfigBuilder(t *testing.T) {
	var b message.ThrottleConfigBuilder

	b.AddRules([]message.ThrottleRule{{Count: 1, Per: 2 * time.Minute}})

	b.WithDestTypes(TestTypeSMS).AddRules([]message.ThrottleRule{{Count: 2, Per: 3 * time.Minute}})

	b.WithMsgTypes(notification.MessageTypeAlert).AddRules([]message.ThrottleRule{{Count: 3, Per: 5 * time.Minute}})

	b.WithDestTypes(TestTypeVoice).WithMsgTypes(notification.MessageTypeTest).AddRules([]message.ThrottleRule{{Count: 5, Per: 7 * time.Minute}})

	cfg := b.Config()

	assert.Equal(t, 7*time.Minute, cfg.MaxDuration())

	check := func(dest nfy.DestType, msg notification.MessageType, expRules []message.ThrottleRule) {
		t.Helper()
		assert.EqualValues(t, expRules, cfg.Rules(message.Message{Type: msg, Dest: nfy.NewDest(dest)}))
	}

	check(
		"",
		notification.MessageTypeUnknown,
		[]message.ThrottleRule{
			{Count: 1, Per: 2 * time.Minute},
		},
	)

	check(
		TestTypeSMS,
		notification.MessageTypeUnknown,
		[]message.ThrottleRule{
			{Count: 1, Per: 2 * time.Minute},
			{Count: 2, Per: 3 * time.Minute},
		},
	)

	check(
		TestTypeVoice,
		notification.MessageTypeAlert,
		[]message.ThrottleRule{
			{Count: 1, Per: 2 * time.Minute},
			{Count: 3, Per: 5 * time.Minute},
		},
	)

	check(
		TestTypeVoice,
		notification.MessageTypeTest,
		[]message.ThrottleRule{
			{Count: 1, Per: 2 * time.Minute},
			{Count: 5, Per: 7 * time.Minute},
		},
	)

	check(
		TestTypeSMS,
		notification.MessageTypeAlert,
		[]message.ThrottleRule{
			{Count: 1, Per: 2 * time.Minute},
			{Count: 2, Per: 3 * time.Minute},
			{Count: 3, Per: 5 * time.Minute},
		},
	)
}
