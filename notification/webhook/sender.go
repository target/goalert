package webhook

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/target/goalert/notification"
)

type Sender struct{}

func NewSender(ctx context.Context) *Sender {
	return &Sender{}
}

// Send will send an for the provided message type.
func (s *Sender) Send(ctx context.Context, msg notification.Message) (*notification.MessageStatus, error) {
	switch m := msg.(type) {
	case notification.Verification:
		// TODO: everything here
		log.Println("CODE", strconv.Itoa(m.Code), "\n\n")
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
