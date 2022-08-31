package mocktwilio

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

type assertSMS struct {
	*assertDev
	Message
}

func (a *assertions) newAssertSMS(baseSMS Message) *assertSMS {
	dev := &assertDev{a, baseSMS.To()}
	sms := &assertSMS{
		assertDev: dev,
		Message:   baseSMS,
	}
	dev.t.Logf("mocktwilio: incoming %s", sms)
	return sms
}

func (sms *assertSMS) String() string {
	return fmt.Sprintf("SMS %s from %s to %s", strconv.Quote(sms.Text()), sms.From(), sms.To())
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
		dev.t.Fatalf("mocktwilio: send SMS %s from %s to %s: %v", strconv.Quote(body), dev.number, dev.AppPhoneNumber, err)
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
		for idx, sms := range dev.messages {
			if !dev.matchMessage(dev.number, keywords, sms) {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), dev.Timeout)
			defer cancel()

			err := sms.SetStatus(ctx, status)
			if err != nil {
				dev.t.Fatalf("mocktwilio: set status '%s' on %s: %v", status, sms, err)
			}

			// remove the message from the list of messages
			dev.messages = append(dev.messages[:idx], dev.messages[idx+1:]...)

			return sms
		}
	}

	dev.refresh()

	t := time.NewTimer(dev.Timeout)
	defer t.Stop()

	ref := time.NewTicker(time.Second)
	defer ref.Stop()

	for {
		select {
		case <-t.C:
			dev.t.Errorf("mocktwilio: timeout after %s waiting for an SMS to %s with keywords: %v", dev.Timeout, dev.number, keywords)
			dev.t.FailNow()
		case <-ref.C:
			dev.refresh()
		case baseSMS := <-dev.Messages():
			sms := dev.newAssertSMS(baseSMS)
			if !dev.matchMessage(dev.number, keywords, sms) {
				dev.messages = append(dev.messages, sms)
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), dev.Timeout)
			defer cancel()

			err := sms.SetStatus(ctx, status)
			if err != nil {
				dev.t.Fatalf("mocktwilio: set status '%s' on %s: %v", status, sms, err)
			}

			return sms
		}
	}
}
