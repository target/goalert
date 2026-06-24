package harness

import (
	"github.com/target/goalert/devtools/mocktwilio"
)

type twilioAssertionDevice struct {
	*twilioAssertionAPIContext
	number string
}

func (dev *twilioAssertionDevice) newMatcher(keywords []string) messageMatcher {
	return messageMatcher{number: dev.number, keywords: keywords}
}

func (dev *twilioAssertionDevice) _expectVoice(keywords ...string) *twilioAssertionVoiceCall {
	dev.t.Helper()
	dev.mx.Lock()
	defer dev.mx.Unlock()

	m := dev.newMatcher(keywords)

	for i, call := range dev.calls {
		if !m.match(call) {
			continue
		}

		dev.calls = append(dev.calls[:i], dev.calls[i+1:]...)

		return &twilioAssertionVoiceCall{twilioAssertionDevice: dev, VoiceCall: call}
	}

	timeout, cancel := dev.triggerTimeout()
	defer cancel()
	for {
		var call *mocktwilio.VoiceCall
		select {
		case call = <-dev.VoiceCalls():
		case msg := <-timeout:
			dev.t.Fatalf("Twilio: timeout after %s waiting for voice call to %s with keywords: %v", msg, dev.formatNumber(dev.number), keywords)
		}
		dev.t.Logf("received voice call to %s: %s", dev.formatNumber(call.To()), call.Body())
		if !m.match(call) {
			dev.calls = append(dev.calls, call)
			continue
		}

		return &twilioAssertionVoiceCall{twilioAssertionDevice: dev, VoiceCall: call}
	}
}

func (dev *twilioAssertionDevice) _expectSMS(includePrev bool, keywords ...string) *twilioAssertionSMS {
	dev.t.Helper()
	dev.mx.Lock()
	defer dev.mx.Unlock()

	m := dev.newMatcher(keywords)

	if includePrev {
		for i, sms := range dev.messages {
			if !m.match(sms) {
				continue
			}

			dev.messages = append(dev.messages[:i], dev.messages[i+1:]...)
			return &twilioAssertionSMS{twilioAssertionDevice: dev, SMS: sms}
		}
	}

	timeout, cancel := dev.triggerTimeout()
	defer cancel()
	for {
		var sms *mocktwilio.SMS
		select {
		case sms = <-dev.SMS():
		case msg := <-timeout:
			dev.t.Fatalf("Twilio: timeout after %s waiting for an SMS to %s with keywords: %v", msg, dev.formatNumber(dev.number), keywords)
		}
		dev.t.Logf("received SMS to %s: %s", dev.formatNumber(sms.To()), sms.Body())
		if !m.match(sms) {
			dev.messages = append(dev.messages, sms)
			continue
		}

		return &twilioAssertionSMS{twilioAssertionDevice: dev, SMS: sms}
	}
}

func (dev *twilioAssertionDevice) ExpectSMS(keywords ...string) ExpectedSMS {
	dev.t.Helper()
	sms := dev._expectSMS(true, keywords...)
	sms.Accept()
	return sms
}

func (dev *twilioAssertionDevice) RejectSMS(keywords ...string) {
	dev.t.Helper()
	sms := dev._expectSMS(true, keywords...)
	sms.Reject()
}

func (sms *twilioAssertionSMS) ThenExpect(keywords ...string) ExpectedSMS {
	sms.t.Helper()
	sms = sms._expectSMS(false, keywords...)
	sms.Accept()
	return sms
}

func (dev *twilioAssertionDevice) ExpectVoice(keywords ...string) ExpectedCall {
	dev.t.Helper()
	call := dev._expectVoice(keywords...)
	dev.mx.Lock()
	call.Accept()
	dev.activeCalls = append(dev.activeCalls, call.VoiceCall)
	dev.mx.Unlock()
	return call
}

func (dev *twilioAssertionDevice) RejectVoice(keywords ...string) {
	dev.t.Helper()
	call := dev._expectVoice(keywords...)
	call.Reject()
}

func (dev *twilioAssertionDevice) SendSMS(body string) {
	dev.t.Helper()
	err := dev.Server.SendSMS(dev.number, dev.sendSMSDest, body)
	if err != nil {
		dev.t.Fatalf("send SMS: from %s: %v", dev.formatNumber(dev.number), err)
	}
}

func (dev *twilioAssertionDevice) IgnoreUnexpectedSMS(keywords ...string) {
	dev.mx.Lock()
	defer dev.mx.Unlock()
	dev.ignoredSMS = append(dev.ignoredSMS, messageMatcher{number: dev.number, keywords: keywords})
}

func (dev *twilioAssertionDevice) IgnoreUnexpectedVoice(keywords ...string) {
	dev.mx.Lock()
	defer dev.mx.Unlock()
	dev.ignoredVoice = append(dev.ignoredVoice, messageMatcher{number: dev.number, keywords: keywords})
}
