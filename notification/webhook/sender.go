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
	AlertID int
	Summary string
	Details string
}

type WebhookAlertBundle struct {
	ServiceID   string
	ServiceName string
	Count       int
}

type WebhookAlertStatus struct {
	AlertID  int
	LogEntry string
}

type WebhookAlertStatusBundle struct {
	LogEntry string
	Count    int
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

	switch m := msg.(type) {
	case notification.Test:
		postWithBody(`{"Message": "Test"}`)
	case notification.Verification:
		postWithBody(`{"Message": "Verification Code: ` + strconv.Itoa(m.Code) + `"}`)
	case notification.Alert:

		var wa WebhookAlert

		jsonbytes, err := json.Marshal(m)
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
		jsonstring := string(jsonbytes)

		postWithBody(jsonstring)

	case notification.AlertBundle:

		var wab WebhookAlertBundle

		jsonbytes, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(jsonbytes, &wab)
		if err != nil {
			return nil, err
		}
		jsonbytes, err = json.Marshal(wab)
		if err != nil {
			return nil, err
		}
		jsonstring := string(jsonbytes)

		postWithBody(jsonstring)

	case notification.AlertStatus:

		var was WebhookAlertStatus

		jsonbytes, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(jsonbytes, &was)
		if err != nil {
			return nil, err
		}
		jsonbytes, err = json.Marshal(was)
		if err != nil {
			return nil, err
		}
		jsonstring := string(jsonbytes)

		postWithBody(jsonstring)

	case notification.AlertStatusBundle:
		var wasb WebhookAlertStatusBundle

		jsonbytes, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(jsonbytes, &wasb)
		if err != nil {
			return nil, err
		}
		jsonbytes, err = json.Marshal(wasb)
		if err != nil {
			return nil, err
		}
		jsonstring := string(jsonbytes)

		postWithBody(jsonstring)

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
