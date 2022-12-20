package twassert

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/target/goalert/devtools/mocktwilio"
)

type sms struct {
	*dev
	mocktwilio.Message
}

func (a *assertions) newAssertSMS(baseSMS mocktwilio.Message) *sms {
	d := &dev{a, baseSMS.To()}
	sms := &sms{
		dev:     d,
		Message: baseSMS,
	}
	d.t.Logf("mocktwilio: incoming %s", sms)
	return sms
}

func (s *sms) String() string {
	return fmt.Sprintf("SMS %s from %s to %s", strconv.Quote(s.Text()), s.From(), s.To())
}

func (s *sms) ThenExpect(keywords ...string) ExpectedSMS {
	s.t.Helper()
	return s._ExpectSMS(false, mocktwilio.MessageDelivered, keywords...)
}

func (s *sms) ThenReply(body string) SMSReply {
	s.SendSMS(body)
	return s
}

func (d *dev) SendSMS(body string) {
	d.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout)
	defer cancel()

	_, err := d.SendMessage(ctx, d.number, d.AppPhoneNumber, body)
	if err != nil {
		d.t.Fatalf("mocktwilio: send SMS %s from %s to %s: %v", strconv.Quote(body), d.number, d.AppPhoneNumber, err)
	}
}

func (d *dev) ExpectSMS(keywords ...string) ExpectedSMS {
	d.t.Helper()
	return d._ExpectSMS(true, mocktwilio.MessageDelivered, keywords...)
}

func (d *dev) RejectSMS(keywords ...string) {
	d.t.Helper()
	d._ExpectSMS(true, mocktwilio.MessageFailed, keywords...)
}

func (d *dev) ThenExpect(keywords ...string) ExpectedSMS {
	d.t.Helper()
	keywords = toLowerSlice(keywords)

	return d._ExpectSMS(false, mocktwilio.MessageDelivered, keywords...)
}

func (d *dev) IgnoreUnexpectedSMS(keywords ...string) {
	d.ignoreSMS = append(d.ignoreSMS, ignoreRule{number: d.number, keywords: keywords})
}

func (d *dev) _ExpectSMS(prev bool, status mocktwilio.FinalMessageStatus, keywords ...string) *sms {
	d.t.Helper()

	keywords = toLowerSlice(keywords)
	if prev {
		for idx, sms := range d.messages {
			if !d.matchMessage(d.number, keywords, sms) {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), d.Timeout)
			defer cancel()

			err := sms.SetStatus(ctx, status)
			if err != nil {
				d.t.Fatalf("mocktwilio: set status '%s' on %s: %v", status, sms, err)
			}

			// remove the message from the list of messages
			d.messages = append(d.messages[:idx], d.messages[idx+1:]...)

			return sms
		}
	}

	d.refresh()

	t := time.NewTimer(d.Timeout)
	defer t.Stop()

	ref := time.NewTicker(time.Second)
	defer ref.Stop()

	for {
		select {
		case <-t.C:
			d.t.Errorf("mocktwilio: timeout after %s waiting for an SMS to %s with keywords: %v", d.Timeout, d.number, keywords)
			d.t.FailNow()
		case <-ref.C:
			d.refresh()
		case baseSMS := <-d.Messages():
			sms := d.newAssertSMS(baseSMS)
			if !d.matchMessage(d.number, keywords, sms) {
				d.messages = append(d.messages, sms)
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), d.Timeout)
			defer cancel()

			err := sms.SetStatus(ctx, status)
			if err != nil {
				d.t.Fatalf("mocktwilio: set status '%s' on %s: %v", status, sms, err)
			}

			return sms
		}
	}
}
