package notification

import (
	"context"
)

type stubSender struct{}

var _ Sender = stubSender{}

func (stubSender) Send(ctx context.Context, msg Message) (string, *Status, error) {
	return "", &Status{State: StateDelivered}, nil
}
