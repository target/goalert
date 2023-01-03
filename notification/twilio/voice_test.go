package twilio

import (
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
)

func TestSpellNumber(t *testing.T) {
	// Test the spell number function
	assert.Equal(t, "1. 2. 3. 4. 5. 6", spellNumber(123456))
}

func TestBuildMessage(t *testing.T) {
	prefix := "This is GoAlert"
	type mockInput struct {
		prefix string
		msg    notification.Message
	}

	testCases := map[string]struct {
		input       mockInput
		expected    *VoiceOptions
		expectedErr error
	}{
		"Test Notification": {
			input: mockInput{
				prefix: prefix,
				msg: notification.Test{
					Dest: notification.Dest{
						ID:    "1",
						Type:  notification.DestTypeVoice,
						Value: "+16125551234",
					},
					CallbackID: "2",
				},
			},
			expected: &VoiceOptions{
				ValidityPeriod: time.Second * 10,
				CallType:       CallTypeTest,
				CallbackParams: url.Values{"msgID": []string{"2"}},
				Params: url.Values{
					"msgBody":      []string{b64enc.EncodeToString([]byte(fmt.Sprintf("%s with a test message.", prefix)))},
					"msgSubjectID": []string{"-1"},
				},
			},
		},
		"AlertBundle Notification": {
			input: mockInput{
				prefix: prefix,
				msg: notification.AlertBundle{
					Dest: notification.Dest{
						ID:    "1",
						Type:  notification.DestTypeVoice,
						Value: "+16125551234",
					},
					CallbackID:  "2",
					ServiceID:   "3",
					ServiceName: "Widget",
					Count:       5,
				},
			},
			expected: &VoiceOptions{
				ValidityPeriod: time.Second * 10,
				CallType:       CallTypeAlert,
				CallbackParams: url.Values{"msgID": []string{"2"}},
				Params: url.Values{
					"msgBody":      []string{b64enc.EncodeToString([]byte(fmt.Sprintf("%s with alert notifications. Service 'Widget' has 5 unacknowledged alerts.", prefix)))},
					"msgBundle":    []string{"1"},
					"msgSubjectID": []string{"-1"},
				},
			},
		},
		"Alert Notification": {
			input: mockInput{
				prefix: prefix,
				msg: notification.Alert{
					Dest: notification.Dest{
						ID:    "1",
						Type:  notification.DestTypeVoice,
						Value: "+16125551234",
					},
					CallbackID: "2",
					AlertID:    3,
					Summary:    "Widget is Broken",
					Details:    "Oh No!",
				},
			},
			expected: &VoiceOptions{
				ValidityPeriod: time.Second * 10,
				CallType:       CallTypeAlert,
				CallbackParams: url.Values{"msgID": []string{"2"}},
				Params: url.Values{
					"msgBody":      []string{b64enc.EncodeToString([]byte(fmt.Sprintf("%s with an alert notification. Widget is Broken.", prefix)))},
					"msgSubjectID": []string{"3"},
				},
			},
		},
		"AlertStatus Notification": {
			input: mockInput{
				prefix: prefix,
				msg: notification.AlertStatus{
					Dest: notification.Dest{
						ID:    "1",
						Type:  notification.DestTypeVoice,
						Value: "+16125551234",
					},
					CallbackID: "2",
					AlertID:    3,
					Summary:    "Widget is Broken",
					Details:    "Oh No!",
					LogEntry:   "Something is Wrong",
				},
			},
			expected: &VoiceOptions{
				ValidityPeriod: time.Second * 10,
				CallType:       CallTypeAlertStatus,
				CallbackParams: url.Values{"msgID": []string{"2"}},
				Params: url.Values{
					"msgBody":      []string{b64enc.EncodeToString([]byte(fmt.Sprintf("%s with a status update for alert 'Widget is Broken'. Something is Wrong", prefix)))},
					"msgSubjectID": []string{"3"},
				},
			},
		},
		"Verification Notification": {
			input: mockInput{
				prefix: prefix,
				msg: notification.Verification{
					Dest: notification.Dest{
						ID:    "1",
						Type:  notification.DestTypeVoice,
						Value: "+16125551234",
					},
					CallbackID: "2",
					Code:       1234,
				},
			},
			expected: &VoiceOptions{
				ValidityPeriod: time.Second * 10,
				CallType:       CallTypeVerify,
				CallbackParams: url.Values{"msgID": []string{"2"}},
				Params: url.Values{
					"msgBody":      []string{b64enc.EncodeToString([]byte(fmt.Sprintf("%s with your 4-digit verification code. The code is: %s. Again, your 4-digit verification code is: %s.", prefix, spellNumber(1234), spellNumber(1234))))},
					"msgSubjectID": []string{"-1"},
				},
			},
		},
		"Bad Type": {
			input: mockInput{
				prefix: prefix,
				msg: notification.ScheduleOnCallUsers{
					Dest: notification.Dest{
						ID:    "1",
						Type:  notification.DestTypeVoice,
						Value: "+16125551234",
					},
					CallbackID:   "2",
					ScheduleID:   "3",
					ScheduleName: "4",
					ScheduleURL:  "5",
				},
			},
			expectedErr: errors.New("unhandled message type: notification.ScheduleOnCallUsers"),
		},
		"Missing prefix": {
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
			expectedErr: errors.New("No prefix provided"),
		},
		"no input": {
			input: mockInput{
				prefix: prefix,
			},
			expectedErr: errors.New("unhandled message type: <nil>"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Arrange / Act
			otps, err := buildMessage(tc.input.prefix, tc.input.msg)

			// Assert
			assert.Equal(t, tc.expected, otps)
			if tc.expectedErr != nil || err != nil {
				// have to do it this way since errors.Errorf will never match due to memory alocations.
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			}
		})
	}
}

func BenchmarkBuildMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = buildMessage(
			fmt.Sprintf("%d", i),
			notification.Test{
				Dest: notification.Dest{
					ID:    "1",
					Type:  notification.DestTypeVoice,
					Value: fmt.Sprintf("+1612555123%d", i),
				},
				CallbackID: "2",
			})
	}
}
