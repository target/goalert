package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/target/goalert/notification"
)

type Sender struct{}

type WebhookAlert struct {
	AlertID int
	Summary string
	Details string
}

func NewSender(ctx context.Context) *Sender {
	return &Sender{}
}

func post(ctx context.Context, urlStr string, v url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", urlStr, bytes.NewBufferString(v.Encode()))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

// Send will send an alert for the provided message type
func (s *Sender) Send(ctx context.Context, msg notification.Message) (*notification.MessageStatus, error) {

	postWithBody := func(body string) (*http.Response, error) {
		v := make(url.Values)
		v.Set("Body", body)
		resp, err := post(ctx, msg.Destination().Value, v)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	switch m := msg.(type) {
	case notification.Test:
		postWithBody("This is a test message from GoAlert")
	case notification.Verification:
		postWithBody(strconv.Itoa(m.Code))
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
