package harness

import (
	"strings"
	"testing"
	"time"

	"github.com/target/goalert/devtools/mocktwilio"
	"github.com/target/goalert/notification/twilio"
)

// TwilioServer is used to assert voice and SMS behavior.
type TwilioServer interface {

	// Device returns a TwilioDevice for the given number.
	//
	// It is safe to call multiple times for the same device.
	Device(number string) TwilioDevice

	// WaitAndAssert will wait for all messages to be processed.
	//
	// It must be called before any calls to Body() will returned.
	//
	// Any unexpected messages (or missing ones) will result in a test failure.
	WaitAndAssert()
}

// A TwilioDevice immitates a device (i.e. a phone) for testing interactions.
type TwilioDevice interface {
	// SendSMS will send a message to GoAlert from the device.
	SendSMS(body string)

	// ExpectSMS will match against an SMS that matches ALL provided keywords (case-insensitive).
	// Each call to ExpectSMS results in the requirement that an additional SMS is received.
	ExpectSMS(keywords ...string) TwilioExpectedMessage

	// ExpectVoice will match against a voice call where the spoken text matches ALL provided keywords (case-insensitive).
	ExpectVoice(keywords ...string) TwilioExpectedCall

	// IgnoreUnexpectedSMS will cause any extra SMS messages (after processing ExpectSMS calls) that match
	// ALL keywords (case-insensitive) to not fail the test.
	IgnoreUnexpectedSMS(keywords ...string)

	// IgnoreUnexpectedVoice will cause any extra voice calls (after processing ExpectVoice) that match
	// ALL keywords (case-insensitive) to not fail the test.
	IgnoreUnexpectedVoice(keywords ...string)
}

// TwilioExpectedCall represents a phone call.
type TwilioExpectedCall interface {
	// ThenPress imitates a user entering a key on the phone.
	ThenPress(digits string) TwilioExpectedCall
	// ThenExpect asserts that the message matches ALL keywords (case-insensitive).
	//
	// Generally used as ThenPress().ThenExpect()
	ThenExpect(keywords ...string) TwilioExpectedCall

	// RespondWithFailed will tell the backend that the call failed.
	RespondWithFailed()

	// Body will return the full spoken message as text. Separate stanzas (e.g. multiple `<Say>`) are
	// separated by newline.
	//
	// WaitAndAssert() must be called first or Body() will hang.
	Body() string
}

// TwilioExpectedMessage represents an SMS message.
type TwilioExpectedMessage interface {

	// ThenReply will respond with an SMS with the given body.
	ThenReply(body string)

	// RespondWithFailed will tell the backend that message delivery failed.
	RespondWithFailed()

	// Body is the text of the SMS message.
	//
	// WaitAndAssert() must be called first or Body() will hang.
	Body() string
}

type matcher interface {
	Match(string) bool
	String() string
}

func matchKeywords(keywords []string) matcher {
	lc := make([]string, len(keywords))
	for i, w := range keywords {
		lc[i] = strings.ToLower(w)
	}
	return keywordMatcher(lc)
}

type joinMatch struct {
	a, b matcher
	and  bool
}

func matchOR(a, b matcher) matcher {
	if _, ok := a.(noneMatch); ok {
		return b
	}
	return &joinMatch{a: a, b: b}
}
func matchAND(a, b matcher) matcher {
	if kw, ok := a.(keywordMatcher); ok && len(kw) == 0 {
		return b
	}
	return &joinMatch{a: a, b: b, and: true}
}
func (j *joinMatch) Match(msg string) bool {
	if j.and {
		return j.a.Match(msg) && j.b.Match(msg)
	}

	return j.a.Match(msg) || j.b.Match(msg)

}
func (j *joinMatch) String() string {
	if j.and {
		return j.a.String() + " AND " + j.b.String()
	}

	return j.a.String() + " OR " + j.b.String()

}

type noneMatch struct{}

func (noneMatch) Match(string) bool { return false }
func (noneMatch) String() string    { return "" }

type keywordMatcher []string

func (k keywordMatcher) String() string {
	return strings.Join([]string(k), ",")
}
func (k keywordMatcher) Match(msg string) bool {
	msg = strings.ToLower(msg)
	for _, word := range k {
		if !strings.Contains(msg, word) {
			return false
		}
	}
	return true
}

type twServer struct {
	*mocktwilio.Server
	t *testing.T
	h *Harness

	devices map[string]*twDevice
}

func newTWServer(t *testing.T, h *Harness, s *mocktwilio.Server) *twServer {
	return &twServer{
		t:       t,
		h:       h,
		Server:  s,
		devices: make(map[string]*twDevice),
	}
}

type expMessage struct {
	dev *twDevice
	matcher
	body  chan string
	fail  bool
	reply string
}
type twDevice struct {
	tw     *twServer
	number string

	ignoreMessages matcher
	ignoreCalls    matcher

	expMessages []*expMessage
	expCalls    []*expCall
}

type expCall struct {
	dev *twDevice
	matcher

	step   int
	body   chan string
	digits string
	fail   bool
	next   *expCall
}

func (dev *twDevice) IgnoreUnexpectedSMS(keywords ...string) {
	dev.ignoreMessages = matchOR(dev.ignoreMessages, matchKeywords(keywords))
}
func (dev *twDevice) IgnoreUnexpectedVoice(keywords ...string) {
	dev.ignoreCalls = matchOR(dev.ignoreCalls, matchKeywords(keywords))
}
func (dev *twDevice) ExpectSMS(keywords ...string) TwilioExpectedMessage {
	msg := &expMessage{
		dev:     dev,
		matcher: matchKeywords(keywords),
		body:    make(chan string, 1),
	}
	dev.expMessages = append(dev.expMessages, msg)
	return msg
}
func (msg *expMessage) ThenReply(body string) {
	msg.reply = body
}
func (msg *expMessage) RespondWithFailed() {
	msg.fail = true
}
func (msg *expMessage) Body() string {
	b := <-msg.body
	msg.body <- b
	return b
}

func (dev *twDevice) ExpectVoice(keywords ...string) TwilioExpectedCall {
	call := &expCall{
		dev:     dev,
		matcher: matchKeywords(keywords),
		body:    make(chan string, 1),
	}
	dev.expCalls = append(dev.expCalls, call)
	return call
}
func (call *expCall) RespondWithFailed() {
	call.fail = true
}
func (call *expCall) ThenExpect(keywords ...string) TwilioExpectedCall {
	call.matcher = matchAND(call.matcher, matchKeywords(keywords))
	return call
}
func (call *expCall) Body() string {
	b := <-call.body
	call.body <- b
	return b
}
func (call *expCall) ThenPress(digits string) TwilioExpectedCall {
	call.digits = digits
	call.next = call.dev.ExpectVoice().(*expCall)
	call.next.step = call.step + 1
	// remove call added by ExpectVoice() since it will be tracked by call.next
	call.dev.expCalls = call.dev.expCalls[:len(call.dev.expCalls)-1]
	return call.next
}

func (tw *twServer) Device(number string) TwilioDevice {
	dev := tw.devices[number]
	if dev != nil {
		return dev
	}
	dev = &twDevice{
		tw:             tw,
		number:         number,
		ignoreMessages: noneMatch{},
		ignoreCalls:    noneMatch{},
	}
	tw.devices[number] = dev
	return dev
}

// Twilio will return the mock Twilio API. It is safe to call multiple times.
func (h *Harness) Twilio() TwilioServer {
	return h.tw
}
func (tw *twServer) timeoutFail() {
	tw.t.Helper()
	for num, dev := range tw.devices {
		for _, msg := range dev.expMessages {
			if msg == nil {
				continue
			}
			tw.t.Errorf("Twilio: Did not receive SMS to %s containing: %s", num, msg.matcher.String())
		}
		for _, call := range dev.expCalls {
			if call == nil {
				continue
			}
			tw.t.Errorf("Twilio: Did not receive voice call to %s (step #%d) containing: %s", num, call.step, call.matcher.String())
		}
	}
	tw.t.Error("Twilio: Timeout after 15 seconds waiting for one or more expected calls/messages.")
}

func (dev *twDevice) SendSMS(body string) {
	dev.tw.SendSMS(dev.number, dev.tw.h.cfg.Twilio.FromNumber, body)
}

func (dev *twDevice) processSMS(sms *mocktwilio.SMS) {
	dev.tw.t.Helper()

	for i, msg := range dev.expMessages {
		if msg == nil {
			continue
		}
		if !msg.Match(sms.Body()) {
			continue
		}

		dev.tw.t.Logf("Twilio: Received expected SMS to %s: %s", sms.To(), sms.Body())

		// matches -- process it
		if msg.fail {
			sms.Reject()
		} else {
			sms.Accept()
		}

		if msg.reply != "" {
			dev.tw.SendSMS(sms.To(), sms.From(), msg.reply)
		}

		msg.body <- sms.Body()
		dev.expMessages[i] = nil
		return
	}

	if dev.ignoreMessages.Match(sms.Body()) {
		// ignored
		return
	}

	// didn't match anything
	dev.tw.unexpectedSMS(sms)
}

func (dev *twDevice) processCall(vc *mocktwilio.VoiceCall) {
	dev.tw.t.Helper()

	if vc.Status() == twilio.CallStatusRinging {
		dev.tw.t.Logf("Twilio: Received voice call to %s, asking for message.", vc.To())
		vc.Accept()
		return
	}

	for i, call := range dev.expCalls {
		if call == nil {
			continue
		}
		msg := vc.Message()
		if !call.Match(msg) {
			continue
		}

		dev.tw.t.Logf("Twilio: Received expected voice call to %s (step #%d): %s", vc.To(), call.step, vc.Message())

		if call.fail {
			vc.Reject()
		}

		if call.next != nil {
			vc.PressDigits(call.digits)
			dev.expCalls[i] = call.next
		} else {
			vc.Hangup()
			dev.expCalls[i] = nil
		}

		call.body <- msg
		return
	}

	// didn't match anything
	dev.tw.unexpectedCall(vc)
}
func (tw *twDevice) done() bool {
	for _, msg := range tw.expMessages {
		if msg != nil {
			return false
		}
	}
	for _, call := range tw.expCalls {
		if call != nil {
			return false
		}
	}
	return true
}
func (tw *twServer) unexpectedSMS(sms *mocktwilio.SMS) {
	tw.t.Helper()
	tw.t.Fatalf("Twilio: Unexpected SMS to %s: %s", sms.To(), sms.Body())
}
func (tw *twServer) unexpectedCall(vc *mocktwilio.VoiceCall) {
	tw.t.Helper()
	if vc.Message() != "" {
		tw.t.Fatalf("Twilio: Unexpected voice call (or message) to %s: %s", vc.To(), vc.Message())
	} else {
		tw.t.Fatalf("Twilio: Unexpected voice call to %s", vc.To())
	}
}
func (tw *twServer) WaitAndAssert() {
	tw.t.Helper()

	processMessages := func() {
		tw.h.Trigger()
		// wait for mock twilio server to send messages
		msgDelay := time.NewTimer(1000 * time.Millisecond)
		defer msgDelay.Stop()
		for {
			select {
			case sms := <-tw.SMS():
				dev := tw.devices[sms.To()]
				if dev == nil {
					tw.unexpectedSMS(sms)
				} else {
					dev.processSMS(sms)
				}
			case vc := <-tw.VoiceCalls():
				dev := tw.devices[vc.To()]
				if dev == nil {
					tw.unexpectedCall(vc)
				} else {
					dev.processCall(vc)
				}
			case err := <-tw.Server.Errors():
				tw.t.Errorf("Twilio: %v", err)
			case <-msgDelay.C:
				return
			}
			msgDelay.Reset(1000 * time.Millisecond)
		}
	}

	processMessages()
	var doneCycles int
	for i := 0; i < 15; i++ {
		var waiting bool
		for _, dev := range tw.devices {
			if !dev.done() {
				waiting = true
				break
			}
		}
		if !waiting {
			doneCycles++
		} else {
			doneCycles = 0
		}

		// complete one extra cycle to check for extra messages
		if doneCycles >= 3 {
			return
		}

		tw.h.FastForward(time.Minute)
		processMessages()
	}

	tw.timeoutFail()
	tw.t.Fatal("Twilio: Did not get all expected messages after 15 cycles.")
}
