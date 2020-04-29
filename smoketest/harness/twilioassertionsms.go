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
	sms.Server.SendSMS(sms.To(), sms.From(), body)
	return sms
}
