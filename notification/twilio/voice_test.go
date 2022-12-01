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

func BenchmarkSpellNumber(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = spellNumber(i)
	}
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
					"msgPause":     []string{"41"},
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
					"msgPause":     []string{"43"},
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
					"msgPause":     []string{"66"},
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
					"msgPause":     []string{"52", "77"},
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
	var msgPauseIndex []int
	for i := 0; i < b.N; i++ {
		msgPauseIndex = append(msgPauseIndex, i)
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

func TestProcessSayBody(t *testing.T) {
	type mockInput struct {
		resp          *twiMLResponse
		msgBody       string
		msgPauseIndex []int
	}

	testCases := map[string]struct {
		input    mockInput
		expected *twiMLResponse
	}{
		"with a pause": {
			input: mockInput{
				resp:          &twiMLResponse{},
				msgBody:       "Hello World!",
				msgPauseIndex: []int{len("Hello")},
			},
			expected: &twiMLResponse{
				say: []sayType{
					{text: "Hello"},
					{pause: true},
					{text: " World!"},
				},
			},
		},
		"legacy with no pause": {
			input: mockInput{
				resp:    &twiMLResponse{},
				msgBody: "Hello World!",
			},
			expected: &twiMLResponse{
				say: []sayType{
					{text: "Hello World!"},
				},
			},
		},
		"no response object": {
			input: mockInput{
				msgBody: "Hello World!",
			},
		},
		"no body": {
			input: mockInput{
				resp: &twiMLResponse{},
			},
			expected: &twiMLResponse{},
		},
		"with a pause at the beginning": {
			input: mockInput{
				resp:          &twiMLResponse{},
				msgBody:       "Hello World!",
				msgPauseIndex: []int{0},
			},
			expected: &twiMLResponse{
				say: []sayType{
					{pause: true},
					{text: "Hello World!"},
				},
			},
		},
		"with a pause at the end": {
			input: mockInput{
				resp:          &twiMLResponse{},
				msgBody:       "Hello World!",
				msgPauseIndex: []int{len("Hello World!")},
			},
			expected: &twiMLResponse{
				say: []sayType{
					{text: "Hello World!"},
					{pause: true},
				},
			},
		},
		"invalid pause index - too big": {
			input: mockInput{
				resp:          &twiMLResponse{},
				msgBody:       "Hello World!",
				msgPauseIndex: []int{len("Hello World!") + 1},
			},
			expected: &twiMLResponse{
				say: []sayType{
					{text: "Hello World!"},
				},
			},
		},
		"invalid pause index - too small": {
			input: mockInput{
				resp:          &twiMLResponse{},
				msgBody:       "Hello World!",
				msgPauseIndex: []int{-1},
			},
			expected: &twiMLResponse{
				say: []sayType{
					{text: "Hello World!"},
				},
			},
		},
		"with a valid and invalid pause index": {
			input: mockInput{
				resp:          &twiMLResponse{},
				msgBody:       "Hello World!",
				msgPauseIndex: []int{len("Hello"), -1},
			},
			expected: &twiMLResponse{
				say: []sayType{
					{text: "Hello"},
					{pause: true},
					{text: " World!"},
				},
			},
		},
		"Several pauses": {
			input: mockInput{
				resp:          &twiMLResponse{},
				msgBody:       "Hello, This is GoAlert with your 4-digit verification code. The code is: 0123. Again, your 4-digit verification code is: 0123.",
				msgPauseIndex: []int{len("Hello, This is GoAlert with your 4-digit verification code."), len("Hello, This is GoAlert with your 4-digit verification code.") + len(" The code is: 0123.")},
			},
			expected: &twiMLResponse{
				say: []sayType{
					{text: "Hello, This is GoAlert with your 4-digit verification code."},
					{pause: true},
					{text: " The code is: 0123."},
					{pause: true},
					{text: " Again, your 4-digit verification code is: 0123."},
				},
			},
		},
		"with a pause with ascii characters": {
			input: mockInput{
				resp:          &twiMLResponse{},
				msgBody:       "你好世界！",
				msgPauseIndex: []int{len("你好")},
			},
			expected: &twiMLResponse{
				say: []sayType{
					{text: "你好"},
					{pause: true},
					{text: "世界！"},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Arrange / Act
			processSayBody(tc.input.resp, tc.input.msgBody, tc.input.msgPauseIndex)

			// Assert
			assert.Equal(t, tc.expected, tc.input.resp)
		})
	}
}

func BenchmarkProcessSayBody(b *testing.B) {
	seed := "Hello World"
	var msgPauseIndex []int
	for i := 0; i < b.N; i++ {
		seed = fmt.Sprintf("%s%d", seed, i)
		msgPauseIndex = append(msgPauseIndex, i)
		processSayBody(&twiMLResponse{}, seed, msgPauseIndex)
	}
}
