package notification

import (
	"context"

	"github.com/pkg/errors"
)

// A Sender can send notifications.
type Sender interface {
	// Send should return nil error if the notification was sent successfully. It should be expected
	// that a returned error means that the notification should be attempted again.
	//
	// If the sent message can have it's status tracked, a unique externalID should be returned.
	Send(context.Context, Message) (externalID string, status *Status, err error)
}

// A StatusChecker is an optional interface a Sender can implement that allows checking the status
// of a previously sent message by it's externalID.
type StatusChecker interface {
	Status(ctx context.Context, externalID string) (*Status, error)
}

// ErrStatusUnsupported should be returned when a Status() check is not supported by the provider.
var ErrStatusUnsupported = errors.New("status check unsupported by provider")

// ReceiverSetter is an optinoal interface a Sender can implement for use with two-way interactions.
type ReceiverSetter interface {
	SetReceiver(Receiver)
}
