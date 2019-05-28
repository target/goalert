package notification

import (
	"context"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type stubSender struct{}

var _ SendResponder = stubSender{}

func (stubSender) Send(ctx context.Context, msg Message) (*MessageStatus, error) {
	return &MessageStatus{
		Ctx:               ctx,
		ProviderMessageID: "stub_" + uuid.NewV4().String(),
		ID:                msg.ID(),
		State:             MessageStateDelivered,
	}, nil
}
func (stubSender) Status(ctx context.Context, id, providerID string) (*MessageStatus, error) {
	return nil, errors.New("not implemented")
}
func (stubSender) ListenStatus() <-chan *MessageStatus     { return make(chan *MessageStatus) }
func (stubSender) ListenResponse() <-chan *MessageResponse { return make(chan *MessageResponse) }
