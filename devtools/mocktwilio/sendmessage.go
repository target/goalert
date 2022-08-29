package mocktwilio

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ttacon/libphonenumber"
)

// SendMessage will send a message to the specified phone number, blocking until delivery.
//
// Messages sent using this method will be treated as "inbound" messages and will not
// appear in Messages().
func (srv *Server) SendMessage(ctx context.Context, from, to, body string) (Message, error) {
	nm, err := libphonenumber.Parse(from, "")
	if err != nil {
		return nil, fmt.Errorf("invalid from phone number '%s': %w", from, err)
	}
	from = libphonenumber.Format(nm, libphonenumber.E164)

	nm, err = libphonenumber.Parse(to, "")
	if err != nil {
		return nil, fmt.Errorf("invalid to phone number '%s': %w", to, err)
	}
	to = libphonenumber.Format(nm, libphonenumber.E164)

	n := srv.number(to)
	if n == nil {
		return nil, fmt.Errorf("unregistered destination number: %s", to)
	}

	if n.SMSWebhookURL == "" {
		return nil, fmt.Errorf("no SMS webhook URL registered for number: %s", to)
	}

	s := srv.newMsgState()
	s.Direction = "inbound"
	s.To = to
	s.From = from
	s.Body = body

	err = s.setFinalStatus(ctx, "received", 0)
	if err != nil {
		return nil, fmt.Errorf("set final status: %w", err)
	}

	v := make(url.Values)
	v.Set("AccountSid", srv.Config().AccountSID)
	v.Set("ApiVersion", "2010-04-01")
	v.Set("Body", body)
	v.Set("From", from)
	// from city/country/state/zip omitted
	v.Set("MessageSid", s.ID)
	v.Set("NumMedia", "0")
	v.Set("NumSegments", "1")
	// SmsMessageSid/SmsSid omitted
	v.Set("SmsStatus", s.Status)
	v.Set("To", to)
	// to city/country/state/zip omitted

	db := <-srv.msgStateDB
	db[s.ID] = s
	srv.msgStateDB <- db

	data, err := srv.post(ctx, n.SMSWebhookURL, v)
	if err != nil {
		return nil, fmt.Errorf("POST to webhook URL %s: %w\n\n%s", n.SMSWebhookURL, err, string(data))
	}

	return &message{s}, nil
}
