package mocktwilio

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/validation/validate"
)

// VoiceCall represents a voice call session.
type VoiceCall struct {
	s *Server

	mx sync.Mutex

	call twilio.Call

	acceptCh chan struct{}
	rejectCh chan struct{}

	messageCh chan string
	pressCh   chan string
	hangupCh  chan struct{}
	doneCh    chan struct{}

	// start is used to track when the call was created (entered queue)
	start time.Time

	// callStart tracks when the call was accepted
	// and is used to calculate call.CallDuration when completed.
	callStart      time.Time
	url            string
	callbackURL    string
	lastMessage    string
	callbackEvents []string
	hangup         bool
}

func (vc *VoiceCall) process() {
	defer vc.s.workers.Done()
	defer close(vc.doneCh)

	if vc.s.wait(vc.s.cfg.MinQueueTime) {
		return
	}

	vc.updateStatus(twilio.CallStatusInitiated)

	if vc.s.wait(vc.s.cfg.MinQueueTime) {
		return
	}

	vc.updateStatus(twilio.CallStatusRinging)

	var err error
	vc.lastMessage, err = vc.fetchMessage("")
	if err != nil {
		vc.s.errs <- fmt.Errorf("fetch message: %w", err)
		return
	}
	select {
	case vc.s.callCh <- vc:
	case <-vc.s.shutdown:
		return
	}

waitForAccept:
	for {
		select {
		case vc.messageCh <- vc.lastMessage:
		case <-vc.acceptCh:
			break waitForAccept
		case <-vc.rejectCh:
			vc.updateStatus(twilio.CallStatusFailed)
			return
		case <-vc.s.shutdown:
			return
		}
	}

	vc.updateStatus(twilio.CallStatusInProgress)
	vc.callStart = time.Now()

	for {
		select {
		case <-vc.rejectCh:
			vc.updateStatus(twilio.CallStatusFailed)
			return
		case <-vc.s.shutdown:
			return
		case <-vc.hangupCh:
			vc.updateStatus(twilio.CallStatusCompleted)
			return
		case vc.messageCh <- vc.lastMessage:
		case digits := <-vc.pressCh:
			vc.lastMessage, err = vc.fetchMessage(digits)
			if err != nil {
				vc.s.errs <- fmt.Errorf("fetch message: %w", err)
				return
			}
			if vc.hangup {
				vc.updateStatus(twilio.CallStatusCompleted)
				return
			}
		}
	}
}

func (s *Server) serveCallStatus(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimSuffix(path.Base(req.URL.Path), ".json")
	vc := s.call(id)

	if vc == nil {
		http.NotFound(w, req)
		return
	}
	err := json.NewEncoder(w).Encode(vc.cloneCall())
	if err != nil {
		panic(err)
	}
}

func (s *Server) call(id string) *VoiceCall {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.calls[id]
}

func (s *Server) serveNewCall(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	vc := VoiceCall{
		acceptCh:  make(chan struct{}),
		doneCh:    make(chan struct{}),
		rejectCh:  make(chan struct{}),
		messageCh: make(chan string),
		pressCh:   make(chan string),
		hangupCh:  make(chan struct{}),
	}

	fromValue := req.FormValue("From")
	s.mx.RLock()
	_, hasCallback := s.callbacks["VOICE:"+fromValue]
	s.mx.RUnlock()
	if !hasCallback {
		apiError(400, w, &twilio.Exception{
			Message: "Wrong from number.",
		})
		return
	}

	vc.s = s
	vc.call.To = req.FormValue("To")
	vc.call.From = fromValue
	vc.call.SID = s.id("CA")
	vc.call.SequenceNumber = new(int)
	vc.callbackURL = req.FormValue("StatusCallback")

	err := validate.URL("StatusCallback", vc.callbackURL)
	if err != nil {
		apiError(400, w, &twilio.Exception{
			Code:    11100,
			Message: err.Error(),
		})
		return
	}
	vc.url = req.FormValue("Url")
	err = validate.URL("StatusCallback", vc.url)
	if err != nil {
		apiError(400, w, &twilio.Exception{
			Code:    11100,
			Message: err.Error(),
		})
		return
	}

	vc.callbackEvents = map[string][]string(req.Form)["StatusCallbackEvent"]
	vc.callbackEvents = append(vc.callbackEvents, "completed", "failed") // always send completed and failed
	vc.start = time.Now()

	vc.call.Status = twilio.CallStatusQueued

	s.mx.Lock()
	s.calls[vc.call.SID] = &vc
	s.mx.Unlock()
	s.callInCh <- &vc

	data, err := json.Marshal(vc.cloneCall())
	if err != nil {
		panic(err)
	}

	w.WriteHeader(201)
	_, err = w.Write(data)
	if err != nil {
		panic(err)
	}
}

func (vc *VoiceCall) updateStatus(stat twilio.CallStatus) {
	// move to queued
	vc.mx.Lock()
	vc.call.Status = stat

	switch stat {
	case twilio.CallStatusInProgress:
		vc.callStart = time.Now()
	case twilio.CallStatusCompleted:
		vc.call.CallDuration = time.Since(vc.callStart)
	}
	*vc.call.SequenceNumber++
	vc.mx.Unlock()

	var sendEvent bool
	evtName := string(stat)
	if evtName == "in-progres" {
		evtName = "answered"
	}
	for _, e := range vc.callbackEvents {
		if e == evtName {
			sendEvent = true
			break
		}
	}

	if !sendEvent {
		return
	}

	// attempt post to status callback
	_, err := vc.s.post(vc.callbackURL, vc.values(""))
	if err != nil {
		vc.s.errs <- errors.Wrap(err, "post to call status callback")
	}
}

func (vc *VoiceCall) values(digits string) url.Values {
	call := vc.cloneCall()

	v := make(url.Values)
	v.Set("CallSid", call.SID)
	v.Set("CallStatus", string(call.Status))
	v.Set("To", call.To)
	v.Set("From", call.From)
	v.Set("Direction", "outbound-api")
	v.Set("SequenceNumber", strconv.Itoa(*call.SequenceNumber))
	if call.Status == twilio.CallStatusCompleted {
		v.Set("CallDuration", strconv.FormatFloat(call.CallDuration.Seconds(), 'f', 1, 64))
	}

	if digits != "" {
		v.Set("Digits", digits)
	}

	return v
}

// VoiceCalls will return a channel that will be fed VoiceCalls as they arrive.
func (s *Server) VoiceCalls() chan *VoiceCall {
	return s.callCh
}

func (vc *VoiceCall) cloneCall() *twilio.Call {
	vc.mx.Lock()
	defer vc.mx.Unlock()

	call := vc.call
	return &call
}

// Accept will allow a call to move from initiated to "in-progress".
func (vc *VoiceCall) Accept() { close(vc.acceptCh) }

// Reject will reject a call, moving it to a "failed" state.
func (vc *VoiceCall) Reject() { close(vc.rejectCh); <-vc.doneCh }

// Hangup will end the call, setting it's state to "completed".
func (vc *VoiceCall) Hangup() { close(vc.hangupCh); <-vc.doneCh }

func (vc *VoiceCall) fetchMessage(digits string) (string, error) {
	data, err := vc.s.post(vc.url, vc.values(digits))
	if err != nil {
		return "", fmt.Errorf("post voice endpoint: %w", err)
	}
	type resp struct {
		XMLName xml.Name `xml:"Response"`
		Say     []string `xml:"Say>prosody"`
		Gather  struct {
			Action string   `xml:"action,attr"`
			Say    []string `xml:"Say>prosody"`
		}
		RedirectURL string    `xml:"Redirect"`
		Hangup      *struct{} `xml:"Hangup"`
	}
	var r resp
	err = xml.Unmarshal(data, &r)
	if err != nil {
		return "", fmt.Errorf("unmarshal XML voice response: %w", err)
	}

	s := append(r.Say, r.Gather.Say...)
	if r.Gather.Action != "" {
		vc.url = r.Gather.Action
	}
	if r.RedirectURL != "" {
		// Twilio's own implementation is totally broken with relative URLs, so we assume absolute (since that's all we use as a consequence)
		vc.url = r.RedirectURL
	}
	if r.Hangup != nil {
		vc.hangup = true
	}

	if r.RedirectURL != "" {
		// redirect and get new message
		return vc.fetchMessage("")
	}

	return strings.Join(s, "\n"), nil
}

// Status will return the current status of the call.
func (vc *VoiceCall) Status() twilio.CallStatus {
	return vc.cloneCall().Status
}

// PressDigits will re-query for a spoken message with the given digits.
func (vc *VoiceCall) PressDigits(digits string) { vc.pressCh <- digits }

// ID returns the unique ID of this phone call.
// It is analogous to the Twilio SID of a call.
func (vc *VoiceCall) ID() string {
	return vc.call.SID
}

// To returns the destination phone number.
func (vc *VoiceCall) To() string {
	return vc.call.To
}

// From return the source phone number.
func (vc *VoiceCall) From() string {
	return vc.call.From
}

// Body will return the last spoken message of the call.
func (vc *VoiceCall) Body() string {
	select {
	case <-vc.doneCh:
		return vc.lastMessage
	case msg := <-vc.messageCh:
		return msg
	case <-vc.s.shutdown:
		return ""
	}
}
