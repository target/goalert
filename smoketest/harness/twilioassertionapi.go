package harness

import (
	"sync"
	"testing"
	"time"

	"github.com/target/goalert/devtools/mocktwilio"
)

type twilioAssertionAPIContext struct {
	t *testing.T
	*twilioAssertionAPI
}

func (w *twilioAssertionAPIContext) Device(number string) PhoneDevice {
	return w.twilioAssertionAPI.Device(w.t, number)
}
func (w *twilioAssertionAPIContext) WaitAndAssert() { w.twilioAssertionAPI.WaitAndAssert(w.t) }

type twilioAssertionAPI struct {
	*mocktwilio.Server

	triggerFn   func()
	sendSMSDest string
	messages    []*mocktwilio.SMS
	calls       []*mocktwilio.VoiceCall

	activeCalls []*mocktwilio.VoiceCall

	ignoredSMS   anyMessage
	ignoredVoice anyMessage

	mx sync.Mutex
}

func newTwilioAssertionAPI(triggerFn func(), srv *mocktwilio.Server, sendSMSDest string) *twilioAssertionAPI {
	return &twilioAssertionAPI{
		triggerFn:   triggerFn,
		Server:      srv,
		sendSMSDest: sendSMSDest,
	}
}

func (tw *twilioAssertionAPI) WithT(t *testing.T) PhoneAssertions {
	return &twilioAssertionAPIContext{t: t, twilioAssertionAPI: tw}
}

func (tw *twilioAssertionAPI) Device(t *testing.T, number string) PhoneDevice {
	return &twilioAssertionDevice{
		twilioAssertionAPIContext: &twilioAssertionAPIContext{t: t, twilioAssertionAPI: tw},

		number: number,
	}
}

func (tw *twilioAssertionAPI) triggerTimeout() (<-chan string, func()) {
	cancelCh := make(chan struct{})

	errMsgCh := make(chan string, 1)
	t := time.NewTimer(15 * time.Second)
	go func() {
		defer t.Stop()
		// 3 engine cycles, or timeout/cancel (whichever is sooner)
		for i := 0; i < 3; i++ {
			select {
			case <-t.C:
				errMsgCh <- "15 seconds"
				return
			default:
				tw.triggerFn()
			}
		}
		errMsgCh <- "3 engine cycles"
	}()

	return errMsgCh, func() { close(cancelCh) }
}

func (tw *twilioAssertionAPI) WaitAndAssert(t *testing.T) {
	t.Helper()
	if t.Failed() {
		// don't wait if test has already failed
		return
	}
	tw.mx.Lock()
	defer tw.mx.Unlock()

	for _, sms := range tw.messages {
		if tw.ignoredSMS.match(sms) {
			continue
		}
		t.Fatalf("got unexpected SMS to %s: %s", sms.To(), sms.Body())
	}
	for _, call := range tw.calls {
		if tw.ignoredVoice.match(call) {
			continue
		}
		t.Fatalf("got unexpected voice call to %s: %s", call.To(), call.Body())
	}

	timeout, cancel := tw.triggerTimeout()
	defer cancel()
waitLoop:
	for {
		select {
		case sms := <-tw.SMS():
			t.Fatalf("got unexpected SMS to %s: %s", sms.To(), sms.Body())
		case call := <-tw.VoiceCalls():
			t.Fatalf("got unexpected voice call to %s: %s", call.To(), call.Body())
		case <-timeout:
			break waitLoop
		}
	}

	tw.ignoredSMS = nil
	tw.ignoredVoice = nil
	tw.calls = nil
	tw.messages = nil
	for _, call := range tw.activeCalls {
		call.Hangup()
	}
	tw.activeCalls = nil
}

// Twilio will return PhoneAssertions for the given testing context.
func (h *Harness) Twilio(t *testing.T) PhoneAssertions { return h.tw.WithT(t) }
