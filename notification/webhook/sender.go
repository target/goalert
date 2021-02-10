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

type WebhookAlert struct {
	AlertID     int    `json:",omitempty"`
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

func getWebhookAlertBody(msg notification.Message) (*string, error) {
	var wa WebhookAlert

	jsonbytes, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonbytes, &wa)
	if err != nil {
		return nil, err
	}
	jsonbytes, err = json.Marshal(wa)
	if err != nil {
		return nil, err
	}

	result := string(jsonbytes)
	return &result, nil
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

	switch m := msg.(type) {
	case notification.Test:
		postWithBody(`{"Message": "Test"}`)
	case notification.Verification:
		postWithBody(`{"Message": "Verification Code: ` + strconv.Itoa(m.Code) + `"}`)
	case notification.Alert, notification.AlertBundle, notification.AlertStatus, notification.AlertStatusBundle:
		body, err := getWebhookAlertBody(m)
		if err != nil {
			return nil, err
		}
		_, err = postWithBody(*body)
		if err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("message type not supported")
	}
	return &notification.MessageStatus{ID: msg.ID(), State: notification.MessageStateSent, ProviderMessageID: msg.ID()}, nil
}

func (s *Sender) Status(ctx context.Context, id, providerID string) (*notification.MessageStatus, error) {
	return nil, errors.New("notification/webhook: status not supported")
}

func (s *Sender) ListenStatus() <-chan *notification.MessageStatus     { return nil }
func (s *Sender) ListenResponse() <-chan *notification.MessageResponse { return nil }
