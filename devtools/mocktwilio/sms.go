package mocktwilio

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"sync"
	"time"
)

type sms struct {
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"date_created"`
	SentAt    *time.Time `json:"date_sent,omitempty"`
	UpdatedAt time.Time  `json:"date_updated"`
	Direction string     `json:"direction"`
	MsgSID    string     `json:"messaging_service_sid,omitempty"`
	ID        string     `json:"sid"`
	Status    string     `json:"status"`
	To        string     `json:"to"`
	From      string     `json:"from"`

	StatusURL string `json:"-"`

	srv      *Server
	mx       sync.Mutex
	setFinal chan FinalMessageStatus
	final    sync.Once
}

func (srv *Server) newSMS() *sms {
	n := time.Now()
	return &sms{
		setFinal:  make(chan FinalMessageStatus, 1),
		CreatedAt: n,
		UpdatedAt: n,
		srv:       srv,
		ID:        srv.nextID("SM"),
	}
}

func (s *sms) lifecycle(ctx context.Context) {
	if s.MsgSID != "" {
		nums := s.srv.msgSvc[s.MsgSID]
		idx := rand.Intn(len(nums))
		newFrom := nums[idx]
		err := s.setSendStatus(ctx, "queued", newFrom.Number)
		if err != nil {
			s.srv.logErr(ctx, err)
		}
	}

	err := s.setSendStatus(ctx, "sending", "")
	if err != nil {
		s.srv.logErr(ctx, err)
	}

	n := s.srv.numbers[s.To]
	if n == nil {
		select {
		case <-ctx.Done():
		case s.srv.messagesCh <- &message{s}:
		}
		return
	}

	// destined for app
	_, err = s.srv.SendMessage(ctx, s.From, s.To, s.Body)
	if err != nil {
		s.srv.logErr(ctx, err)
		err = s.setFinalStatus(ctx, "undelivered", 30006)
	} else {
		err = s.setFinalStatus(ctx, "delivered", 0)
	}
	if err != nil {
		s.srv.logErr(ctx, err)
	}
}

func (s *sms) setSendStatus(ctx context.Context, status, updateFrom string) error {
	s.mx.Lock()
	if updateFrom != "" {
		s.From = updateFrom
	}
	s.Status = status
	s.UpdatedAt = time.Now()
	if status == "sent" {
		s.SentAt = &s.UpdatedAt
	}
	s.mx.Unlock()

	if s.Direction == "inbound" {
		return nil
	}
	if s.StatusURL == "" {
		return nil
	}

	v := make(url.Values)
	v.Set("AccountSid", s.srv.cfg.AccountSID)
	v.Set("ApiVersion", "2010-04-01")
	v.Set("From", s.From)
	v.Set("MessageSid", s.ID)
	v.Set("MessageStatus", s.Status)
	// SmsSid/SmsStatus omitted
	v.Set("To", s.To)

	_, err := s.srv.post(ctx, s.StatusURL, v)
	if err != nil {
		return fmt.Errorf("send status callback: %v", err)
	}

	return nil
}

func (s *sms) setFinalStatus(ctx context.Context, status FinalMessageStatus, code int) error {
	var err error
	s.final.Do(func() {
		err = s.setSendStatus(ctx, string(status), "")
	})

	return err
}

func (s *sms) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}

	type data sms
	s.mx.Lock()
	defer s.mx.Unlock()
	return json.Marshal((*data)(s))
}
