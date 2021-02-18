package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/target/goalert/notification"
)

type Sender struct{}

type POSTData struct {
	AlertID     int    `json:",omitempty"`
	Type        string `json:",omitempty"`
	Code        string `json:",omitempty"`
	Summary     string `json:",omitempty"`
	Details     string `json:",omitempty"`
	ServiceID   string `json:",omitempty"`
	ServiceName string `json:",omitempty"`
	Count       int    `json:",omitempty"`
	LogEntry    string `json:",omitempty"`
}

func NewSender(ctx context.Context) *Sender {
	return &Sender{}
}

// Send will send an alert for the provided message type
func (s *Sender) Send(ctx context.Context, msg notification.Message) (*notification.MessageStatus, error) {

	postWithBody := func(body string) (*http.Response, error) {
		req, err := http.NewRequest("POST", msg.Destination().Value, bytes.NewBufferString(body))
		if err != nil {
			return nil, err
		}
		req = req.WithContext(ctx)
		req.Header.Add("Content-Type", "application/json")
		return http.DefaultClient.Do(req)
	}

	var payload POSTData

	switch m := msg.(type) {
	case notification.Test:
		payload.Type = "Test"
		payload.Summary = "Test Message"
		payload.Details = "This is a test message from GoAlert"

	case notification.Verification:
		payload.Type = "Verification"
		payload.Summary = "Verification Message"
		payload.Details = "This is a verification message from GoAlert"
		payload.Code = strconv.Itoa(m.Code)

	case notification.Alert:
		payload.Type = "Alert"
		payload.AlertID = m.AlertID
		payload.Summary = m.Summary
		payload.Details = m.Details

	case notification.AlertBundle:
		payload.Type = "AlertBundle"
		payload.ServiceID = m.ServiceID
		payload.ServiceName = m.ServiceName
		payload.Count = m.Count

	case notification.AlertStatus:
		payload.Type = "AlertStatus"
		payload.AlertID = m.AlertID
		payload.LogEntry = m.LogEntry

	case notification.AlertStatusBundle:
		payload.Type = "AlertStatusBundle"
		payload.AlertID = m.AlertID
		payload.Count = m.Count
		payload.LogEntry = m.LogEntry

	default:
		return nil, errors.New("message type not supported")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	postWithBody(string(data))

	return &notification.MessageStatus{ID: msg.ID(), State: notification.MessageStateSent, ProviderMessageID: msg.ID()}, nil
}

func (s *Sender) Status(ctx context.Context, id, providerID string) (*notification.MessageStatus, error) {
	return nil, errors.New("notification/webhook: status not supported")
}

func (s *Sender) ListenStatus() <-chan *notification.MessageStatus     { return nil }
func (s *Sender) ListenResponse() <-chan *notification.MessageResponse { return nil }
