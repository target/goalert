package mocktwilio

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/validation/validate"
)

// VoiceCall represents a voice call session.
type VoiceCall struct {
	s *Server

	call twilio.Call

	// start is used to track when the call was created (entered queue)
	start time.Time

	// callStart tracks when the call was accepted
	// and is used to cacluate call.CallDuration when completed.
	callStart       time.Time
	url             string
	callbackURL     string
	callbackEvents  []string
	message         string
	needsProcessing bool
	hangup          bool
}

func (s *Server) serveCallStatus(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimSuffix(path.Base(req.URL.Path), ".json")
	var call twilio.Call

	s.mx.RLock()
	vc := s.calls[id]
	if vc != nil {
		call = vc.call // copy while we have the read lock
	}
	s.mx.RUnlock()

	if vc == nil {
		http.NotFound(w, req)
		return
	}
	err := json.NewEncoder(w).Encode(call)
	if err != nil {
		panic(err)
	}

}

func (s *Server) serveNewCall(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	var vc VoiceCall

	vc.call.From = req.FormValue("From")
	if s.callbacks["VOICE:"+vc.call.From] == "" {
		apiError(400, w, &twilio.Exception{
			Message: "Wrong from number.",
		})
		return
	}
	vc.s = s
	vc.call.To = req.FormValue("To")
	vc.call.SID = s.id("CA")
	vc.call.SequenceNumber = new(int)
	vc.callbackURL = req.FormValue("StatusCallback")
	err := validate.URL("StatusCallback", vc.callbackURL)
	if err != nil {
		apiError(400, w, &twilio.Exception{
			Code:    11100,
			Message: err.Error(),
		})
	}
	vc.url = req.FormValue("Url")
	err = validate.URL("StatusCallback", vc.url)
	if err != nil {
		apiError(400, w, &twilio.Exception{
			Code:    11100,
			Message: err.Error(),
		})
	}

	vc.callbackEvents = map[string][]string(req.Form)["StatusCallbackEvent"]
	vc.callbackEvents = append(vc.callbackEvents, "completed", "failed") // always send completed and failed
	vc.start = time.Now()

	vc.call.Status = twilio.CallStatusQueued

	s.mx.Lock()
	s.calls[vc.call.SID] = &vc
	s.mx.Unlock()

	w.WriteHeader(201)
	err = json.NewEncoder(w).Encode(vc.call)
	if err != nil {
		panic(err)
	}

}

func (vc *VoiceCall) updateStatus(stat twilio.CallStatus) {
	// move to queued
	vc.s.mx.Lock()
	vc.call.Status = stat
	if stat == twilio.CallStatusInProgress {
		vc.callStart = time.Now()
	}
	if stat == twilio.CallStatusCompleted {
		vc.call.CallDuration = time.Since(vc.callStart)
	}
	*vc.call.SequenceNumber++
	vc.s.mx.Unlock()
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
	vc.s.mx.RLock()
	v := make(url.Values)
	v.Set("CallSid", vc.call.SID)
	v.Set("CallStatus", string(vc.call.Status))
	v.Set("To", vc.call.To)
	v.Set("From", vc.call.From)
	v.Set("Direction", "outbound-api")
	v.Set("SequenceNumber", strconv.Itoa(*vc.call.SequenceNumber))
	if vc.call.Status == twilio.CallStatusCompleted {
		v.Set("CallDuration", strconv.FormatFloat(vc.call.CallDuration.Seconds(), 'f', 1, 64))
	}

	if digits != "" {
		v.Set("Digits", digits)
	}
	vc.s.mx.RUnlock()

	return v
}

// VoiceCalls will return a channel that will be fed VoiceCalls as they arrive.
func (s *Server) VoiceCalls() chan *VoiceCall {
	return s.callCh
}

// Accept will allow a call to move from initiated to "in-progress".
func (vc *VoiceCall) Accept() {
	vc.updateStatus(twilio.CallStatusInProgress)
	vc.PressDigits("")
}

// Reject will reject a call, moving it to a "failed" state.
func (vc *VoiceCall) Reject() {
	vc.updateStatus(twilio.CallStatusFailed)
}

// Hangup will end the call, setting it's state to "completed".
func (vc *VoiceCall) Hangup() {
	vc.updateStatus(twilio.CallStatusCompleted)
}

// PressDigits will re-query for a spoken message with the given digits.
//
// It also causes the result of Listen() to be blank until a new message is gathered.
func (vc *VoiceCall) PressDigits(digits string) {
	data, err := vc.s.post(vc.url, vc.values(digits))
	if err != nil {
		vc.s.errs <- err
		return
	}
	type resp struct {
		XMLName xml.Name `xml:"Response"`
		Say     []string `xml:"Say"`
		Gather  struct {
			Action string   `xml:"action,attr"`
			Say    []string `xml:"Say"`
		}
		RedirectURL string    `xml:"Redirect"`
		Hangup      *struct{} `xml:"Hangup"`
	}
	var r resp
	err = xml.Unmarshal(data, &r)
	if err != nil {
		vc.s.errs <- errors.Wrap(err, "unmarshal XML voice response")
		return
	}

	// use data to update callbackURL and/or message
	s := append(r.Say, r.Gather.Say...)
	vc.s.mx.Lock()
	if r.Gather.Action != "" {
		vc.url = r.Gather.Action
	}
	if r.RedirectURL != "" {
		vc.needsProcessing = false
		// Twilio's own implementation is totally broken with relative URLs, so we assume absolute (since that's all we use as a consequence)
		vc.url = r.RedirectURL
	} else {
		vc.needsProcessing = true
	}
	if r.Hangup != nil {
		vc.hangup = true
	}
	vc.message = strings.Join(s, "\n")
	vc.s.mx.Unlock()

	if r.RedirectURL != "" {
		// redirect and get new message
		vc.PressDigits("")
	}
}

// ID returns the unique ID of this phone call.
// It is analogus to the Twilio SID of a call.
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

// Message will return the last spoken message of the call.
func (vc *VoiceCall) Message() string {
	vc.s.mx.RLock()
	defer vc.s.mx.RUnlock()
	return vc.message
}
