package notification

import (
	"context"

	"github.com/target/goalert/auth/authlink"
	"github.com/target/goalert/notification/nfy"
)

// A ResultReceiver processes notification responses.
type ResultReceiver interface {
	SetSendResult(ctx context.Context, res *SendResult) error

	Receive(ctx context.Context, callbackID string, result Result) error
	ReceiveSubject(ctx context.Context, providerID, subjectID, callbackID string, result Result) error
	AuthLinkURL(ctx context.Context, providerID, subjectID string, meta authlink.Metadata) (string, error)
	Start(context.Context, nfy.Dest) error
	Stop(context.Context, nfy.Dest) error

	IsKnownDest(ctx context.Context, dest nfy.Dest) (bool, error)
}
