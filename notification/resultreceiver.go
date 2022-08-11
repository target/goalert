package notification

import (
	"context"
	"net/url"
)

// A ResultReceiver processes notification responses.
type ResultReceiver interface {
	SetSendResult(ctx context.Context, res *SendResult) error

	Receive(ctx context.Context, callbackID string, result Result) error
	ReceiveSubject(ctx context.Context, providerID, subjectID, callbackID string, result Result) error
	AuthLinkURL(ctx context.Context, providerID, subjectID string, params url.Values) (string, error)
	Start(context.Context, Dest) error
	Stop(context.Context, Dest) error

	IsKnownDest(ctx context.Context, destType DestType, destValue string) (bool, error)
}
