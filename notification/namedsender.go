package notification

import (
	"context"
)

type namedSender struct {
	Sender
	name     string
	destType string
}

func (s *namedSender) Send(ctx context.Context, msg Message) (*SendResult, error) {
	sent, err := s.Sender.Send(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &SendResult{
		ID: msg.MsgID(),
		Status: Status{
			State:    sent.State,
			Details:  sent.StateDetails,
			SrcValue: sent.SrcValue,
		},
		ProviderMessageID: ProviderMessageID{
			ProviderName: s.name,
			ExternalID:   sent.ExternalID,
		},
	}, nil
}
