package notification

import (
	"github.com/pkg/errors"
)

// ErrStatusUnsupported should be returned when a Status() check is not supported by the provider.
var ErrStatusUnsupported = errors.New("status check unsupported by provider")

// ErrStatusNotFound should be returned when a Status() check fails because the provider no longer
// has a record of the message (e.g., it has expired or been purged upstream)
var ErrStatusNotFound = errors.New("message not found by provider")

// ReceiverSetter is an optional interface a Sender can implement for use with two-way interactions.
type ReceiverSetter interface {
	SetReceiver(Receiver)
}
