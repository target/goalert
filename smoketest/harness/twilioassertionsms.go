package harness

import (
	"github.com/target/goalert/devtools/mocktwilio"
)

type twilioAssertionSMS struct {
	*twilioAssertionDevice
	*mocktwilio.SMS
}

var _ ExpectedSMS = &twilioAssertionSMS{}

func (sms *twilioAssertionSMS) ThenReply(body string) SMSReply {
	// TODO: error here with T context
	err := sms.Server.SendSMS(sms.To(), sms.From(), body)
	if err != nil {
		sms.t.Fatalf("send SMS: from %s: %v", sms.formatNumber(sms.To()), err)
	}
	return sms
}
