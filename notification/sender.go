package notification

import (
	"context"

	"github.com/pkg/errors"
	"github.com/target/goalert/notification/nfymsg"
)

// A Sender can send notifications.
type Sender interface {
	// Send should return nil error if the notification was sent successfully. It should be expected
	// that a returned error means that the notification should be attempted again.
	//
	// If the sent message can have its status tracked, a unique externalID should be returned.
	Send(context.Context, Message) (*nfymsg.SentMessage, error)
}

// A StatusChecker is an optional interface a Sender can implement that allows checking the status
// of a previously sent message by it's externalID.
type StatusChecker interface {
	Status(ctx context.Context, externalID string) (*Status, error)
}

// A FriendlyValuer is an optional interface a Sender can implement that
// allows retrieving a friendly name for a destination value.
//
// For example, a formatted phone number or username for a Slack ID.
type FriendlyValuer interface {
	FriendlyValue(context.Context, string) (string, error)
}

// ErrStatusUnsupported should be returned when a Status() check is not supported by the provider.
var ErrStatusUnsupported = errors.New("status check unsupported by provider")

// ReceiverSetter is an optional interface a Sender can implement for use with two-way interactions.
type ReceiverSetter interface {
	SetReceiver(Receiver)
}
