package notification

import (
	"context"
)

type stubSender struct{}

var _ Sender = stubSender{}

func (stubSender) Send(ctx context.Context, msg Message) (*SentMessage, error) {
	return &SentMessage{State: StateDelivered}, nil
}
