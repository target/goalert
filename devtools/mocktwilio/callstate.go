package mocktwilio

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"
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

	action   chan struct{}
	lastResp string
}

func (srv *Server) newCallState() *callState {
	n := time.Now()
	return &callState{
		srv:       srv,
		ID:        srv.nextID("CA"),
		CreatedAt: n,
		UpdatedAt: n,
		action:    make(chan struct{}, 1),
	}
}

func (s *callState) Text() string {
	s.mx.Lock()
	defer s.mx.Unlock()
	return s.lastResp
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

	return nil
}

func (s *callState) status() string {
	s.mx.Lock()
	defer s.mx.Unlock()
	return s.Status
}

func (s *callState) lifecycle(ctx context.Context) {
}

func (s *callState) setStatus(ctx context.Context, status string) {
	s.mx.Lock()
	defer s.mx.Unlock()
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

	// TODO: post status, pass to logErr
}

func (s *callState) Press(ctx context.Context, key string) error {
	s.action <- struct{}{}
	if s.status() != "in-progress" {
		<-s.action
		return fmt.Errorf("call not in progress")
	}

	err := s.update(ctx, "")
	if err != nil {
		s.setStatus(ctx, "failed")
		<-s.action
		return err
	}

	<-s.action
	return nil
}

func (s *callState) End(ctx context.Context, status FinalCallStatus) error {
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
