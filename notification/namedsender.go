package notification

import (
	"context"
)

type namedSender struct {
	Sender
	name     string
	destType DestType
}

func (s *namedSender) Send(ctx context.Context, msg Message) (*MessageStatus, error) {
	status, err := s.Sender.Send(ctx, msg)
	return status.wrap(ctx, s), err
}
