package notification

//go:generate go run golang.org/x/tools/cmd/stringer -type Result

import (
	"context"
)

// Result specifies a response to a notification.
type Result int

// Possible notification responses.
const (
	ResultAcknowledge Result = iota
	ResultResolve
	ResultStart
	ResultStop
)

// A Receiver is something that can process a notification result.
type Receiver interface {
	UpdateStatus(context.Context, *MessageStatus) error
	Receive(ctx context.Context, callbackID string, result Result) error
	Start(context.Context, Dest) error
	Stop(context.Context, Dest) error
}

// A Sender is something that can send a notification.
type Sender interface {

	// Send should return nil if the notification was sent successfully. It should be expected
	// that a returned error means that the notification should be attempted again.
	Send(context.Context, Message) (*MessageStatus, error)
}

// A StatusChecker allows checking the status of a sent message.
type StatusChecker interface {
	Status(ctx context.Context, messageID, providerMessageID string) (*MessageStatus, error)
}

// StatusUpdater will provide status updates for messages.
type StatusUpdater interface {
	ListenStatus() <-chan *MessageStatus
}

// A Responder will provide message responses.
type Responder interface {
	ListenResponse() <-chan *MessageResponse
}

// MessageResponse represents a received response from a user.
type MessageResponse struct {
	Ctx    context.Context
	ID     string
	From   Dest
	Result Result

	Err chan error
}
