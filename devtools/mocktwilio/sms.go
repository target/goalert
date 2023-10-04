package mocktwilio

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/validation/validate"
)

// SMS represents an SMS message.
type SMS struct {
	s         *Server
	msg       twilio.Message
	body      string
	statusURL string
	destURL   string
	start     time.Time
	mx        sync.Mutex

	acceptCh chan bool
	doneCh   chan struct{}
}

func (s *Server) sendSMS(fromValue, to, body, statusURL, destURL string) (*SMS, error) {
	fromNumber := s.getFromNumber(fromValue)
	if statusURL != "" {
		err := validate.URL("StatusCallback", statusURL)
		if err != nil {
			return nil, twilio.Exception{
				Code:    11100,
				Message: err.Error(),
			}
		}
		s.mx.RLock()
		_, hasCallback := s.callbacks["SMS:"+fromNumber]
		s.mx.RUnlock()

		if !hasCallback {
			return nil, twilio.Exception{
				Code:    21606,
				Message: `The "From" phone number provided is not a valid, SMS-capable inbound phone number for your account.`,
			}
		}
	}
	if destURL != "" {
		err := validate.URL("Callback", destURL)
		if err != nil {
			return nil, twilio.Exception{
				Code:    11100,
				Message: err.Error(),
			}
		}
	}

	sms := &SMS{
		s: s,
		msg: twilio.Message{
			To:     to,
			Status: twilio.MessageStatusAccepted,
			SID:    s.id("SM"),
		},
		statusURL: statusURL,
		destURL:   destURL,
		start:     time.Now(),
		body:      body,
		acceptCh:  make(chan bool, 1),
		doneCh:    make(chan struct{}),
	}

	if strings.HasPrefix(fromValue, "MG") {
		sms.msg.MessagingServiceSID = fromValue
	} else {
		sms.msg.From = fromValue
	}

	s.mx.Lock()
	s.messages[sms.msg.SID] = sms
	s.mx.Unlock()

	s.smsInCh <- sms

	return sms, nil
}

func (s *Server) serveNewMessage(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	sms, err := s.sendSMS(req.FormValue("From"), req.FormValue("To"), req.FormValue("Body"), req.FormValue("StatusCallback"), "")

	if e := (twilio.Exception{}); errors.As(err, &e) {
		apiError(400, w, &e)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(sms.cloneMessage())
	if err != nil {
		panic(err)
	}

	w.WriteHeader(201)
	_, err = w.Write(data)
	if err != nil {
		panic(err)
	}
}

func (s *Server) serveMessageStatus(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimSuffix(path.Base(req.URL.Path), ".json")
	sms := s.sms(id)
	if sms == nil {
		http.NotFound(w, req)
		return
	}

	err := json.NewEncoder(w).Encode(sms.cloneMessage())
	if err != nil {
		panic(err)
	}
}

func (sms *SMS) updateStatus(stat twilio.MessageStatus) {
	sms.mx.Lock()
	sms.msg.Status = stat
	switch stat {
	case twilio.MessageStatusAccepted, twilio.MessageStatusQueued:
	default:
		if sms.msg.MessagingServiceSID == "" {
			break
		}

		sms.msg.From = sms.s.getFromNumber(sms.msg.MessagingServiceSID)
	}
	sms.mx.Unlock()

	if sms.statusURL == "" {
		return
	}

	// attempt post to status callback
	_, err := sms.s.post(sms.statusURL, sms.values(false))
	if err != nil {
		sms.s.errs <- err
	}
}
func (sms *SMS) cloneMessage() *twilio.Message {
	sms.mx.Lock()
	defer sms.mx.Unlock()

	msg := sms.msg
	return &msg
}

func (s *Server) sms(id string) *SMS {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.messages[id]
}

func (sms *SMS) values(body bool) url.Values {
	v := make(url.Values)
	msg := sms.cloneMessage()
	v.Set("MessageStatus", string(msg.Status))
	v.Set("MessageSid", msg.SID)
	v.Set("To", msg.To)
	v.Set("From", msg.From)
	if body {
		v.Set("Body", sms.body)
	}
	return v
}

// SMS will return a channel that will be fed incoming SMS messages as they arrive.
func (s *Server) SMS() chan *SMS {
	return s.smsCh
}

// SendSMS will cause an SMS to be sent to the given number with the contents of body.
//
// The to parameter must match a value passed to RegisterSMSCallback or an error is returned.
func (s *Server) SendSMS(from, to, body string) error {
	s.mx.RLock()
	cbURL := s.callbacks["SMS:"+to]
	s.mx.RUnlock()

	if cbURL == "" {
		return fmt.Errorf(`unknown/unregistered destination (to) number "%s"`, to)
	}

	sms, err := s.sendSMS(from, to, body, "", cbURL)
	if err != nil {
		return err
	}

	<-sms.doneCh

	return nil
}

func (sms *SMS) process() {
	defer sms.s.workers.Done()
	defer close(sms.doneCh)

	if sms.s.wait(sms.s.cfg.MinQueueTime) {
		return
	}

	sms.updateStatus(twilio.MessageStatusQueued)

	if sms.s.wait(sms.s.cfg.MinQueueTime) {
		return
	}

	sms.updateStatus(twilio.MessageStatusSending)

	if sms.destURL != "" {
		// inbound SMS
		_, err := sms.s.post(sms.destURL, sms.values(true))
		if err != nil {
			sms.s.errs <- err
			sms.updateStatus(twilio.MessageStatusUndelivered)
		} else {
			sms.updateStatus(twilio.MessageStatusDelivered)
		}
		return
	}

	select {
	case <-sms.s.shutdown:
		return
	case sms.s.smsCh <- sms:
	}

	select {
	case <-sms.s.shutdown:
		return
	case accepted := <-sms.acceptCh:
		if accepted {
			sms.updateStatus(twilio.MessageStatusDelivered)
		} else {
			sms.updateStatus(twilio.MessageStatusFailed)
		}
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
	sms.acceptCh <- true
	close(sms.acceptCh)
}

// Reject will cause the SMS to be marked as failed.
func (sms *SMS) Reject() {
	sms.acceptCh <- false
	close(sms.acceptCh)
}
