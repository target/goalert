package notification

import (
	"context"

	"github.com/target/goalert/alert"
)

// A ResultReceiver processes notification responses.
type ResultReceiver interface {
	SetSendResult(ctx context.Context, res *SendResult) error

	Receive(ctx context.Context, callbackID string, result Result) (*alert.Alert, error)
	ReceiveFor(ctx context.Context, callbackID, providerID, subjectID string, result Result) (*alert.Alert, error)

	Start(context.Context, Dest) error
	Stop(context.Context, Dest) error

	IsKnownDest(ctx context.Context, destType DestType, destValue string) (bool, error)
}
