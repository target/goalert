package mocktwilio

import (
	"context"
)

type Message interface {
	ID() string

	To() string
	From() string
	Text() string

	// SetStatus will set the final status of the message.
	SetStatus(context.Context, FinalMessageStatus) error
}

type FinalMessageStatus string

const (
	MessageSent        FinalMessageStatus = "sent"
	MessageDelivered   FinalMessageStatus = "delivered"
	MessageUndelivered FinalMessageStatus = "undelivered"
	MessageFailed      FinalMessageStatus = "failed"
	MessageReceived    FinalMessageStatus = "received"
)

// Messages returns a channel of outbound messages.
func (srv *Server) Messages() <-chan Message { return srv.messagesCh }

type message struct {
	*sms
}

func (msg *message) ID() string   { return msg.sms.ID }
func (msg *message) To() string   { return msg.sms.To }
func (msg *message) From() string { return msg.sms.From }
func (msg *message) Text() string { return msg.sms.Body }

func (msg *message) SetStatus(ctx context.Context, status FinalMessageStatus) error {
	return msg.setFinalStatus(ctx, status, 0)
}
