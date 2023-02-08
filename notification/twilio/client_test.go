package twilio

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
)

func TestSetMsgParams(t *testing.T) {
	testCases := map[string]struct {
		input       notification.Message
		expected    VoiceOptions
		expectedErr error
	}{
		"Test Notification": {
			input: notification.Test{
				Dest: notification.Dest{
					ID:    "1",
					Type:  notification.DestTypeVoice,
					Value: "+16125551234",
				},
				CallbackID: "2",
			},
			expected: VoiceOptions{
				CallType:       "test",
				CallbackParams: url.Values{"msgID": []string{"2"}},
				Params:         url.Values{"msgSubjectID": []string{"-1"}},
			},
		},
		"AlertBundle Notification": {
			input: notification.AlertBundle{
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
			expected: VoiceOptions{
				CallType:       "alert",
				CallbackParams: url.Values{"msgID": []string{"2"}},
				Params: url.Values{
					"msgBundle":    []string{"1"},
					"msgSubjectID": []string{"-1"},
				},
			},
		},
		"Alert Notification": {
			input: notification.Alert{
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
			expected: VoiceOptions{
				CallType:       "alert",
				CallbackParams: url.Values{"msgID": []string{"2"}},
				Params:         url.Values{"msgSubjectID": []string{"3"}},
			},
		},
		"AlertStatus Notification": {
			input: notification.AlertStatus{
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
			expected: VoiceOptions{
				CallType:       "alert-status",
				CallbackParams: url.Values{"msgID": []string{"2"}},
				Params:         url.Values{"msgSubjectID": []string{"3"}},
			},
		},
		"Verification Notification": {
			input: notification.Verification{
				Dest: notification.Dest{
					ID:    "1",
					Type:  notification.DestTypeVoice,
					Value: "+16125551234",
				},
				CallbackID: "2",
				Code:       1234,
			},
			expected: VoiceOptions{
				CallType:       "verify",
				CallbackParams: url.Values{"msgID": []string{"2"}},
				Params:         url.Values{"msgSubjectID": []string{"-1"}},
			},
		},
		"Bad Type": {
			input: notification.ScheduleOnCallUsers{
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
			expected: VoiceOptions{
				CallbackParams: url.Values{},
				Params:         url.Values{},
			},
			expectedErr: errors.New("unhandled message type: notification.ScheduleOnCallUsers"),
		},
		"no input": {
			expected: VoiceOptions{
				CallbackParams: url.Values{},
				Params:         url.Values{},
			},
			expectedErr: errors.New("unhandled message type: <nil>"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Arrange / Act
			result := &VoiceOptions{}
			err := result.setMsgParams(tc.input)

			// Assert
			assert.Equal(t, tc.expected, *result)
			if tc.expectedErr != nil || err != nil {
				// have to do it this way since errors.Errorf will never match due to memory alocations.
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestSetMsgBody(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected VoiceOptions
	}{
		"Test Notification": {
			input: "This is GoAlert with a test message.",
			expected: VoiceOptions{
				Params: url.Values{"msgBody": []string{b64enc.EncodeToString([]byte("This is GoAlert with a test message."))}},
			},
		},
		"no input": {
			expected: VoiceOptions{
				Params: url.Values{"msgBody": []string{b64enc.EncodeToString([]byte(""))}},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Arrange / Act
			result := &VoiceOptions{}
			result.setMsgBody(tc.input)

			// Assert
			assert.Equal(t, tc.expected, *result)
		})
	}
}
