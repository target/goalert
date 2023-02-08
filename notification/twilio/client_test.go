package twilio

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
)

func TestSetMsgParams(t *testing.T) {
	prefix := "This is GoAlert"
	type mockInput struct {
		msg notification.Message
	}

	testCases := map[string]struct {
		input       mockInput
		expected    VoiceOptions
		expectedErr error
	}{
		"Test Notification": {
			input: mockInput{
				msg: notification.Test{
					Dest: notification.Dest{
						ID:    "1",
						Type:  notification.DestTypeVoice,
						Value: "+16125551234",
					},
					CallbackID: "2",
				},
			},
			expected: fmt.Sprintf("%s with a test message.", prefix),
		},
		// "AlertBundle Notification": {
		// 	input: mockInput{
		// 		prefix: prefix,
		// 		msg: notification.AlertBundle{
		// 			Dest: notification.Dest{
		// 				ID:    "1",
		// 				Type:  notification.DestTypeVoice,
		// 				Value: "+16125551234",
		// 			},
		// 			CallbackID:  "2",
		// 			ServiceID:   "3",
		// 			ServiceName: "Widget",
		// 			Count:       5,
		// 		},
		// 	},
		// 	expected: fmt.Sprintf("%s with alert notifications. Service 'Widget' has 5 unacknowledged alerts.", prefix),
		// },
		// "Alert Notification": {
		// 	input: mockInput{
		// 		prefix: prefix,
		// 		msg: notification.Alert{
		// 			Dest: notification.Dest{
		// 				ID:    "1",
		// 				Type:  notification.DestTypeVoice,
		// 				Value: "+16125551234",
		// 			},
		// 			CallbackID: "2",
		// 			AlertID:    3,
		// 			Summary:    "Widget is Broken",
		// 			Details:    "Oh No!",
		// 		},
		// 	},
		// 	expected: fmt.Sprintf("%s with an alert notification. Widget is Broken.", prefix),
		// },
		// "AlertStatus Notification": {
		// 	input: mockInput{
		// 		prefix: prefix,
		// 		msg: notification.AlertStatus{
		// 			Dest: notification.Dest{
		// 				ID:    "1",
		// 				Type:  notification.DestTypeVoice,
		// 				Value: "+16125551234",
		// 			},
		// 			CallbackID: "2",
		// 			AlertID:    3,
		// 			Summary:    "Widget is Broken",
		// 			Details:    "Oh No!",
		// 			LogEntry:   "Something is Wrong",
		// 		},
		// 	},
		// 	expected: fmt.Sprintf("%s with a status update for alert 'Widget is Broken'. Something is Wrong", prefix),
		// },
		// "Verification Notification": {
		// 	input: mockInput{
		// 		prefix: prefix,
		// 		msg: notification.Verification{
		// 			Dest: notification.Dest{
		// 				ID:    "1",
		// 				Type:  notification.DestTypeVoice,
		// 				Value: "+16125551234",
		// 			},
		// 			CallbackID: "2",
		// 			Code:       1234,
		// 		},
		// 	},
		// 	expected: fmt.Sprintf("%s with your 4-digit verification code. The code is: %s. Again, your 4-digit verification code is: %s.", prefix, spellNumber(1234), spellNumber(1234)),
		// },
		// "Bad Type": {
		// 	input: mockInput{
		// 		prefix: prefix,
		// 		msg: notification.ScheduleOnCallUsers{
		// 			Dest: notification.Dest{
		// 				ID:    "1",
		// 				Type:  notification.DestTypeVoice,
		// 				Value: "+16125551234",
		// 			},
		// 			CallbackID:   "2",
		// 			ScheduleID:   "3",
		// 			ScheduleName: "4",
		// 			ScheduleURL:  "5",
		// 		},
		// 	},
		// 	expectedErr: errors.New("unhandled message type: notification.ScheduleOnCallUsers"),
		// },
		// "Missing prefix": {
		// 	input: mockInput{
		// 		msg: notification.Test{
		// 			Dest: notification.Dest{
		// 				ID:    "1",
		// 				Type:  notification.DestTypeVoice,
		// 				Value: "+16125551234",
		// 			},
		// 			CallbackID: "2",
		// 		},
		// 	},
		// 	expectedErr: errors.New("buildMessage error: no prefix provided"),
		// },
		// "no input": {
		// 	input: mockInput{
		// 		prefix: prefix,
		// 	},
		// 	expectedErr: errors.New("unhandled message type: <nil>"),
		// },
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Arrange / Act
			result := &VoiceOptions{}
			err := result.setMsgParams(tc.input.msg)

			// Assert
			assert.Equal(t, tc.expected, *result)
			if tc.expectedErr != nil || err != nil {
				// have to do it this way since errors.Errorf will never match due to memory alocations.
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			}
		})
	}
}

// func BenchmarkBuildMessage(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		_, _ = buildMessage(
// 			fmt.Sprintf("%d", i),
// 			notification.Test{
// 				Dest: notification.Dest{
// 					ID:    "1",
// 					Type:  notification.DestTypeVoice,
// 					Value: fmt.Sprintf("+1612555123%d", i),
// 				},
// 				CallbackID: "2",
// 			})
// 	}
// }
