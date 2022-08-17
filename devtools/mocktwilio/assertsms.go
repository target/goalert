package mocktwilio

import (
	"context"
	"time"
)

type assertSMS struct {
	*assertDev
	Message
}

func (sms *assertSMS) ThenExpect(keywords ...string) ExpectedSMS {
	sms.t.Helper()
	return sms._ExpectSMS(false, MessageDelivered, keywords...)
}

func (sms *assertSMS) ThenReply(body string) SMSReply {
	sms.SendSMS(body)
	return sms
}

func (dev *assertDev) SendSMS(body string) {
	dev.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), dev.Timeout)
	defer cancel()

	_, err := dev.SendMessage(ctx, dev.number, dev.AppPhoneNumber, body)
	if err != nil {
		dev.t.Fatalf("mocktwilio: send SMS %s to %s: %v", body, dev.number, err)
	}
}

func (dev *assertDev) ExpectSMS(keywords ...string) ExpectedSMS {
	dev.t.Helper()
	return dev._ExpectSMS(true, MessageDelivered, keywords...)
}

func (dev *assertDev) RejectSMS(keywords ...string) {
	dev.t.Helper()
	dev._ExpectSMS(true, MessageFailed, keywords...)
}

func (dev *assertDev) ThenExpect(keywords ...string) ExpectedSMS {
	dev.t.Helper()
	keywords = toLowerSlice(keywords)

	return dev._ExpectSMS(false, MessageDelivered, keywords...)
}

func (dev *assertDev) IgnoreUnexpectedSMS(keywords ...string) {
	dev.ignoreSMS = append(dev.ignoreSMS, assertIgnore{number: dev.number, keywords: keywords})
}

func (dev *assertDev) _ExpectSMS(prev bool, status FinalMessageStatus, keywords ...string) *assertSMS {
	dev.t.Helper()

	keywords = toLowerSlice(keywords)
	if prev {
		for _, msg := range dev.messages {
			if !dev.matchMessage(dev.number, keywords, msg) {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), dev.Timeout)
			defer cancel()

			err := msg.SetStatus(ctx, status)
			if err != nil {
				dev.t.Fatalf("mocktwilio: error setting SMS status %s to %s: %v", status, msg.To(), err)
			}

			dev.t.Log("mocktwilio: received expected SMS from", msg.From(), "to", msg.To(), "with text", msg.Text())
			return &assertSMS{assertDev: dev, Message: msg}
		}
	}

	dev.refresh()

	t := time.NewTimer(dev.Timeout)
	defer t.Stop()

	for {
		select {
		case <-t.C:

			dev.t.Errorf("mocktwilio: timeout after %s waiting for an SMS to %s with keywords: %v", dev.Timeout, dev.number, keywords)
			for i, msg := range dev.messages {
				dev.t.Errorf("mocktwilio: message %d: from=%s; to=%s; text=%s", i, msg.From(), msg.To(), msg.Text())
			}

			dev.t.FailNow()
		case msg := <-dev.Messages():
			if !dev.matchMessage(dev.number, keywords, msg) {
				dev.messages = append(dev.messages, msg)
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), dev.Timeout)
			defer cancel()

			err := msg.SetStatus(ctx, status)
			if err != nil {
				dev.t.Fatalf("mocktwilio: error setting SMS status %s to %s: %v", status, msg.To(), err)
			}

			dev.t.Log("mocktwilio: received expected SMS from", msg.From(), "to", msg.To(), "with text", msg.Text())
			return &assertSMS{assertDev: dev, Message: msg}
		}
	}
}
