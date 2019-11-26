package mocktwilio

import (
	"encoding/json"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/validation/validate"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func (s *Server) serveNewMessage(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	var sms SMS
	sms.msg.From = req.FormValue("From")
	if s.callbacks["SMS:"+sms.msg.From] == "" {
		apiError(400, w, &twilio.Exception{
			Code:    21606,
			Message: `The "From" phone number provided is not a valid, SMS-capable inbound phone number for your account.`,
		})
		return
	}
	sms.msg.To = req.FormValue("To")
	sms.msg.SID = s.id("SM")
	sms.msg.Status = twilio.MessageStatusAccepted
	sms.callbackURL = req.FormValue("StatusCallback")
	sms.start = time.Now()
	sms.s = s
	err := validate.URL("StatusCallback", sms.callbackURL)
	if err != nil {
		apiError(400, w, &twilio.Exception{
			Code:    11100,
			Message: err.Error(),
		})
	}

	sms.body = req.FormValue("Body")

	s.mx.Lock()
	s.messages[sms.msg.SID] = &sms
	s.mx.Unlock()

	w.WriteHeader(201)
	err = json.NewEncoder(w).Encode(sms.msg)
	if err != nil {
		panic(err)
	}
}
func (s *Server) serveMessageStatus(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimSuffix(path.Base(req.URL.Path), ".json")
	var msg twilio.Message

	s.mx.RLock()
	sms := s.messages[id]
	if sms != nil {
		msg = sms.msg // copy while we have the read lock
	}
	s.mx.RUnlock()

	if sms == nil {
		http.NotFound(w, req)
		return
	}
	err := json.NewEncoder(w).Encode(msg)
	if err != nil {
		panic(err)
	}
}

// SMS represents an SMS message.
type SMS struct {
	s           *Server
	msg         twilio.Message
	body        string
	callbackURL string
	start       time.Time
}

func (sms *SMS) updateStatus(stat twilio.MessageStatus) {
	// move to queued
	sms.s.mx.Lock()
	sms.msg.Status = stat
	sms.s.mx.Unlock()

	// attempt post to status callback
	_, err := sms.s.post(sms.callbackURL, sms.values(false))
	if err != nil {
		sms.s.errs <- errors.Wrap(err, "post to SMS status callback")
	}
}

func (sms *SMS) values(body bool) url.Values {
	v := make(url.Values)
	sms.s.mx.RLock()
	v.Set("MessageStatus", string(sms.msg.Status))
	v.Set("MessageSid", sms.msg.SID)
	v.Set("To", sms.msg.To)
	v.Set("From", sms.msg.From)
	if body {
		v.Set("Body", sms.body)
	}
	sms.s.mx.RUnlock()
	return v
}

// SMS will return a channel that will be fed incomming SMS messages as they arrive.
func (s *Server) SMS() chan *SMS {
	return s.smsCh
}

// SendSMS will cause an SMS to be sent to the given number with the contents of body.
//
// The to parameter must match a value passed to RegisterSMSCallback or an error is returned.
func (s *Server) SendSMS(from, to, body string) {
	s.mx.Lock()
	cbURL := s.callbacks["SMS:"+to]
	s.mx.Unlock()

	if cbURL == "" {
		s.errs <- errors.New("unknown/unregistered desination (to) number")
	}

	v := make(url.Values)
	v.Set("From", from)
	v.Set("Body", body)

	_, err := s.post(cbURL, v)
	if err != nil {
		s.errs <- err
	}
}

// ID will return the unique ID for this SMS.
func (sms *SMS) ID() string {
	return sms.msg.SID
}

// From returns the phone number the SMS was sent from.
func (sms *SMS) From() string {
	return sms.msg.From
}

// To returns the phone number the SMS is being sent to.
func (sms *SMS) To() string {
	return sms.msg.To
}

// Body returns the contents of the SMS message.
func (sms *SMS) Body() string {
	return sms.body
}

// Accept will cause the SMS to be marked as delivered.
func (sms *SMS) Accept() {
	sms.updateStatus(twilio.MessageStatusDelivered)
}

// Reject will cause the SMS to be marked as undelivered (failed).
func (sms *SMS) Reject() {
	sms.updateStatus(twilio.MessageStatusFailed)
}
