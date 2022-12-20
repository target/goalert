package mocktwilio

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/target/goalert/devtools/mocktwilio/twiml"
)

type callState struct {
	CreatedAt time.Time  `json:"date_created"`
	UpdatedAt time.Time  `json:"date_updated"`
	StartedAt *time.Time `json:"start_time,omitempty"`
	EndedAt   *time.Time `json:"end_time,omitempty"`
	Duration  string     `json:"duration,omitempty"`
	Direction string     `json:"direction"`
	MsgSID    string     `json:"messaging_service_sid,omitempty"`
	ID        string     `json:"sid"`
	Status    string     `json:"status"`
	To        string     `json:"to"`
	From      string     `json:"from"`

	StatusURL string `json:"-"`
	CallURL   string `json:"-"`

	srv *Server
	mx  sync.Mutex
	seq int

	action chan struct{}

	run *twiml.Interpreter

	text string
}

func (srv *Server) newCallState() *callState {
	n := time.Now()
	return &callState{
		srv:       srv,
		ID:        srv.nextID("CA"),
		CreatedAt: n,
		Status:    "queued",
		UpdatedAt: n,
		action:    make(chan struct{}, 1),
		run:       twiml.NewInterpreter(),
	}
}

func (s *callState) lifecycle(ctx context.Context) {
	s.setStatus(ctx, "initiated")
	s.setStatus(ctx, "ringing")

	select {
	case <-ctx.Done():
	case s.srv.callCh <- &call{s}:
	}
}

// Text returns the last spoken text of the call.
func (s *callState) Text() string {
	s.mx.Lock()
	defer s.mx.Unlock()
	return s.text
}

func (s *callState) IsActive() bool {
	switch s.status() {
	case "completed", "failed", "busy", "no-answer", "canceled":
		return false
	}

	return true
}

func (s *callState) Answer(ctx context.Context) error {
	s.action <- struct{}{}
	if s.status() != "ringing" {
		<-s.action
		return nil
	}

	s.setStatus(ctx, "in-progress")

	err := s.update(ctx, "")
	if err != nil {
		s.setStatus(ctx, "failed")
		<-s.action
		return err
	}

	<-s.action
	return nil
}

func (s *callState) update(ctx context.Context, digits string) error {
	v := make(url.Values)
	v.Set("AccountSid", s.srv.cfg.AccountSID)
	v.Set("ApiVersion", "2010-04-01")
	v.Set("CallSid", s.ID)
	v.Set("CallStatus", "in-progress")
	v.Set("Direction", s.Direction)
	v.Set("From", s.From)
	v.Set("To", s.To)
	if digits != "" {
		v.Set("Digits", digits)
	}

	data, err := s.srv.post(ctx, s.CallURL, v)
	if err != nil {
		return err
	}

	err = s.run.SetResponse(data)
	if err != nil {
		return err
	}

	return s.process(ctx)
}

// process will interpret the returned TwiML and update the call state and spoken text.
func (s *callState) process(ctx context.Context) error {
	s.text = ""

	for s.run.Next() {
		switch t := s.run.Verb().(type) {
		case *twiml.Say:
			s.text += t.Content + "\n"
		case *twiml.Redirect:
			err := s.setCallURL(t.URL)
			if err != nil {
				return err
			}
			return s.update(ctx, "")
		case *twiml.Gather:
			return nil
		case *twiml.Reject:
			s.setStatus(ctx, t.Reason)
			return nil
		case *twiml.Hangup:
			s.setStatus(ctx, "completed")
			return nil
		case *twiml.Pause:
			// ignored
		}
	}

	return nil
}

func (s *callState) setCallURL(url string) error {
	newURL := relURL(s.CallURL, url)
	if newURL == "" {
		return fmt.Errorf("invalid redirect url: %s", url)
	}

	s.CallURL = newURL
	return nil
}

func (s *callState) status() string {
	s.mx.Lock()
	defer s.mx.Unlock()
	return s.Status
}

// setStatus will update the call status, posting to the status callback URL, if provided.
func (s *callState) setStatus(ctx context.Context, status string) {
	s.mx.Lock()
	defer s.mx.Unlock()
	if s.Status == status {
		return
	}

	s.Status = status
	s.UpdatedAt = time.Now()
	switch status {
	case "in-progress":
		s.StartedAt = &s.UpdatedAt
	case "completed", "failed", "busy", "no-answer", "canceled":
		s.EndedAt = &s.UpdatedAt
		if s.StartedAt == nil {
			s.StartedAt = &s.UpdatedAt
		}
	}

	if s.StatusURL == "" {
		return
	}

	v := make(url.Values)
	v.Set("AccountSid", s.srv.cfg.AccountSID)
	v.Set("ApiVersion", "2010-04-01")
	if s.StartedAt != nil {
		dur := time.Since(*s.StartedAt)
		if s.EndedAt != nil {
			dur = s.EndedAt.Sub(*s.StartedAt)
		}
		v.Set("Duration", strconv.Itoa(int(math.Ceil(dur.Seconds()))))
	}
	v.Set("CallSid", s.ID)
	v.Set("CallStatus", status)
	v.Set("Direction", s.Direction)
	v.Set("From", s.From)
	v.Set("To", s.To)
	s.seq++
	v.Set("SequenceNumber", strconv.Itoa(s.seq))

	_, err := s.srv.post(ctx, s.StatusURL, v)
	if err != nil {
		s.srv.logErr(ctx, err)
	}
}

func (s *callState) Press(ctx context.Context, digits string) error {
	s.action <- struct{}{}
	if s.status() != "in-progress" {
		<-s.action
		return fmt.Errorf("call not in progress")
	}

	g, ok := s.run.Verb().(*twiml.Gather)
	if !ok {
		<-s.action
		return fmt.Errorf("gather not in progress")
	}
	err := s.setCallURL(g.Action)
	if err != nil {
		s.setStatus(ctx, "failed")
		<-s.action
		return err
	}

	err = s.update(ctx, digits)
	if err != nil {
		s.setStatus(ctx, "failed")
		<-s.action
		return err
	}

	<-s.action
	return nil
}

func (s *callState) PressTimeout(ctx context.Context) error {
	s.action <- struct{}{}
	if s.status() != "in-progress" {
		<-s.action
		return fmt.Errorf("call not in progress")
	}

	g, ok := s.run.Verb().(*twiml.Gather)
	if !ok {
		<-s.action
		return fmt.Errorf("gather not in progress")
	}
	if !g.ActionOnEmptyResult {
		err := s.process(ctx)
		<-s.action
		return err
	}

	err := s.setCallURL(g.Action)
	if err != nil {
		s.setStatus(ctx, "failed")
		<-s.action
		return err
	}

	err = s.update(ctx, "")
	if err != nil {
		s.setStatus(ctx, "failed")
		<-s.action
		return err
	}

	<-s.action
	return nil
}

func (s *callState) Hangup(ctx context.Context, status FinalCallStatus) error {
	s.action <- struct{}{}
	if !s.IsActive() {
		<-s.action
		return nil
	}

	s.setStatus(ctx, string(status))
	return nil
}

func (v *callState) MarshalJSON() ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}

	v.mx.Lock()
	defer v.mx.Unlock()

	if v.StartedAt != nil && v.EndedAt != nil {
		v.Duration = strconv.Itoa(int(v.EndedAt.Sub(*v.StartedAt).Seconds()))
	} else {
		v.Duration = ""
	}

	type data callState
	return json.Marshal((*data)(v))
}
