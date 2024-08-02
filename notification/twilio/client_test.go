package twilio

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfymsg"
)

func TestSetMsgParams(t *testing.T) {
	t.Run("Test Notification", func(t *testing.T) {
		result := &VoiceOptions{}
		err := result.setMsgParams(
			notification.Test{
				Base: nfymsg.Base{
					MsgDest: NewVoiceDest("+16125551234"),
					MsgID:   "2",
				},
			},
		)
		expected := VoiceOptions{
			CallType:       "test",
			CallbackParams: url.Values{"msgID": []string{"2"}},
			Params:         url.Values{"msgSubjectID": []string{"-1"}},
		}

		assert.Equal(t, expected, *result)
		assert.NoError(t, err)
	})
	t.Run("AlertBundle Notification", func(t *testing.T) {
		result := &VoiceOptions{}
		err := result.setMsgParams(
			notification.AlertBundle{
				Base: nfymsg.Base{
					MsgDest: NewVoiceDest("+16125551234"),
					MsgID:   "2",
				},
				ServiceID:   "3",
				ServiceName: "Widget",
				Count:       5,
			},
		)
		expected := VoiceOptions{
			CallType:       "alert",
			CallbackParams: url.Values{"msgID": []string{"2"}},
			Params: url.Values{
				"msgBundle":    []string{"1"},
				"msgSubjectID": []string{"-1"},
			},
		}

		assert.Equal(t, expected, *result)
		assert.NoError(t, err)
	})
	t.Run("Alert Notification", func(t *testing.T) {
		result := &VoiceOptions{}
		err := result.setMsgParams(
			notification.Alert{
				Base: nfymsg.Base{
					MsgDest: NewVoiceDest("+16125551234"),
					MsgID:   "2",
				},
				AlertID: 3,
				Summary: "Widget is Broken",
				Details: "Oh No!",
			},
		)
		expected := VoiceOptions{
			CallType:       "alert",
			CallbackParams: url.Values{"msgID": []string{"2"}},
			Params:         url.Values{"msgSubjectID": []string{"3"}},
		}

		assert.Equal(t, expected, *result)
		assert.NoError(t, err)
	})
	t.Run("AlertStatus Notification", func(t *testing.T) {
		result := &VoiceOptions{}
		err := result.setMsgParams(
			notification.AlertStatus{
				Base: nfymsg.Base{
					MsgDest: NewVoiceDest("+16125551234"),
					MsgID:   "2",
				},
				AlertID:  3,
				Summary:  "Widget is Broken",
				Details:  "Oh No!",
				LogEntry: "Something is Wrong",
			},
		)
		expected := VoiceOptions{
			CallType:       "alert-status",
			CallbackParams: url.Values{"msgID": []string{"2"}},
			Params:         url.Values{"msgSubjectID": []string{"3"}},
		}

		assert.Equal(t, expected, *result)
		assert.NoError(t, err)
	})
	t.Run("Verification Notification", func(t *testing.T) {
		result := &VoiceOptions{}
		err := result.setMsgParams(
			notification.Verification{
				Base: nfymsg.Base{
					MsgDest: NewVoiceDest("+16125551234"),
					MsgID:   "2",
				},
				Code: "1234",
			},
		)
		expected := VoiceOptions{
			CallType:       "verify",
			CallbackParams: url.Values{"msgID": []string{"2"}},
			Params:         url.Values{"msgSubjectID": []string{"-1"}},
		}

		assert.Equal(t, expected, *result)
		assert.NoError(t, err)
	})
	t.Run("Bad Type", func(t *testing.T) {
		result := &VoiceOptions{}
		err := result.setMsgParams(
			notification.ScheduleOnCallUsers{
				Base: nfymsg.Base{
					MsgDest: NewVoiceDest("+16125551234"),
					MsgID:   "2",
				},
				ScheduleID:   "3",
				ScheduleName: "4",
				ScheduleURL:  "5",
			},
		)
		expected := VoiceOptions{
			CallbackParams: url.Values{},
			Params:         url.Values{},
		}

		assert.Equal(t, expected, *result)
		assert.Equal(t, err.Error(), "unhandled message type: nfymsg.ScheduleOnCallUsers")
	})
	t.Run("no input", func(t *testing.T) {
		result := &VoiceOptions{}
		err := result.setMsgParams(nil)
		expected := VoiceOptions{
			CallbackParams: url.Values{},
			Params:         url.Values{},
		}

		assert.Equal(t, expected, *result)
		assert.Equal(t, err.Error(), "unhandled message type: <nil>")
	})
}

func TestSetMsgBody(t *testing.T) {
	t.Run("Test Notification", func(t *testing.T) {
		result := &VoiceOptions{}
		result.setMsgBody("This is GoAlert with a test message.")
		expected := &VoiceOptions{
			Params: url.Values{"msgBody": []string{b64enc.EncodeToString([]byte("This is GoAlert with a test message."))}},
		}
		assert.Equal(t, expected, result)
	})
	t.Run("no input", func(t *testing.T) {
		result := &VoiceOptions{}
		result.setMsgBody("")
		expected := &VoiceOptions{
			Params: url.Values{"msgBody": []string{b64enc.EncodeToString([]byte(""))}},
		}
		assert.Equal(t, expected, result)
	})
}
