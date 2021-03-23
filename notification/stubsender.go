package notification

import (
	"context"

	uuid "github.com/satori/go.uuid"
)

type stubSender struct{}

var _ Sender = stubSender{}

func (stubSender) Send(ctx context.Context, msg Message) (*MessageStatus, error) {
	return &MessageStatus{
		Ctx:               ctx,
		ProviderMessageID: "stub_" + uuid.NewV4().String(),
		ID:                msg.ID(),
		State:             MessageStateDelivered,
	}, nil
}
