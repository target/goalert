package harness

import (
	"github.com/target/goalert/devtools/mocktwilio"
)

type twilioAssertionVoiceCall struct {
	*twilioAssertionDevice
	*mocktwilio.VoiceCall
}

var _ ExpectedCall = &twilioAssertionVoiceCall{}

func (call *twilioAssertionVoiceCall) ThenExpect(keywords ...string) ExpectedCall {
	call.t.Helper()
	msg := call.Body()
	if !containsAllIgnoreCase(msg, keywords) {
		call.t.Fatalf("voice call message from %s was '%s'; expected keywords: %v", call.From(), msg, keywords)
	}

	return call
}
func (call *twilioAssertionVoiceCall) ThenPress(digits string) ExpectedCall {
	call.PressDigits(digits)
	return call
}
func (call *twilioAssertionVoiceCall) Hangup() {
	call.mx.Lock()
	defer call.mx.Unlock()
	call.VoiceCall.Hangup()

	for i, ac := range call.activeCalls {
		if ac == call.VoiceCall {
			call.activeCalls = append(call.activeCalls[:i], call.activeCalls[i+1:]...)
			break
		}
	}
}
