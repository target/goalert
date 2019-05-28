package notification

import (
	"context"
)

type namedSender struct {
	SendResponder
	name     string
	destType DestType
}

func (s *namedSender) Send(ctx context.Context, msg Message) (*MessageStatus, error) {
	status, err := s.SendResponder.Send(ctx, msg)
	return status.wrap(ctx, s), err
}
