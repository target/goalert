package notification

import (
	"context"
)

type namedSender struct {
	Sender
	name     string
	destType DestType
}

func (s *namedSender) Send(ctx context.Context, msg Message) (*SendResult, error) {
	externalID, status, err := s.Sender.Send(ctx, msg)
	if err != nil {
		return nil, err
	}

	res := &SendResult{
		ID: msg.ID(),
	}
	if status != nil {
		res.Status = *status
	}
	res.ProviderMessageID.ProviderName = s.name
	res.ProviderMessageID.ExternalID = externalID

	return res, err
}
