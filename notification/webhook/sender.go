package webhook

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/target/goalert/notification"
)

type Sender struct{}

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

// Send will send an for the provided message type.
func (s *Sender) Send(ctx context.Context, msg notification.Message) (*notification.MessageStatus, error) {

	switch m := msg.(type) {
	case notification.Verification:
		webhookURL := msg.Destination().Value
		v := make(url.Values)
		v.Set("Body", strconv.Itoa(m.Code))
		_, err := post(ctx, webhookURL, v)
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
