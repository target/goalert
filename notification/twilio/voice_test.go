package twilio

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfymsg"
)

func TestSpellNumber(t *testing.T) {
	// Test the spell number function
	assert.Equal(t, "1. 2. 3. 4. 5. 6", spellCode("123456"))
}

func TestBuildMessage(t *testing.T) {
	prefix := "This is GoAlert"

	// Test Notification
	result, err := buildMessage(
		prefix,
		notification.Test{},
	)
	assert.Equal(t, fmt.Sprintf("%s with a test message.", prefix), result)
	assert.NoError(t, err)

	// AlertBundle Notification
	result, err = buildMessage(
		prefix,
		notification.AlertBundle{
			Base:        nfymsg.Base{ID: "2"},
			ServiceID:   "3",
			ServiceName: "Widget",
			Count:       5,
		},
	)
	assert.Equal(t, fmt.Sprintf("%s with alert notifications. Service 'Widget' has 5 unacknowledged alerts.", prefix), result)
	assert.NoError(t, err)

	// Alert Notification
	result, err = buildMessage(
		prefix,
		notification.Alert{
			Base:    nfymsg.Base{ID: "2"},
			AlertID: 3,
			Summary: "Widget is Broken",
			Details: "Oh No!",
		},
	)
	assert.Equal(t, fmt.Sprintf("%s with an alert notification. Widget is Broken.", prefix), result)
	assert.NoError(t, err)

	// AlertStatus Notification
	result, err = buildMessage(
		prefix,
		notification.AlertStatus{
			Base:     nfymsg.Base{ID: "2"},
			AlertID:  3,
			Summary:  "Widget is Broken",
			Details:  "Oh No!",
			LogEntry: "Something is Wrong",
		},
	)
	assert.Equal(t, fmt.Sprintf("%s with a status update for alert 'Widget is Broken'. Something is Wrong", prefix), result)
	assert.NoError(t, err)

	// Verification Notification
	result, err = buildMessage(
		prefix,
		notification.Verification{
			Base: nfymsg.Base{ID: "2"},
			Code: "1234",
		},
	)
	assert.Equal(t, fmt.Sprintf("%s with your 4-digit verification code. The code is: %s. Again, your 4-digit verification code is: %s.", prefix, spellCode("1234"), spellCode("1234")), result)
	assert.NoError(t, err)

	// Bad Type
	result, err = buildMessage(
		prefix,
		notification.ScheduleOnCallUsers{
			Base:         nfymsg.Base{ID: "2"},
			ScheduleID:   "3",
			ScheduleName: "4",
			ScheduleURL:  "5",
		},
	)
	assert.Empty(t, result)
	assert.Error(t, err)

	// Missing prefix
	result, err = buildMessage(
		"",
		notification.Test{},
	)
	assert.Empty(t, result)
	assert.Error(t, err)

	// no input
	result, err = buildMessage(
		prefix,
		nil,
	)
	assert.Empty(t, result)
	assert.Error(t, err)
}

func BenchmarkBuildMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = buildMessage(
			fmt.Sprintf("%d", i),
			notification.Test{
				Base: nfymsg.Base{
					Dest: NewVoiceDest(fmt.Sprintf("+1612555123%d", i)),
					ID:   "2",
				},
			},
		)
	}
}
