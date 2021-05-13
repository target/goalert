package notification

import (
	"context"
)

// A Processor processes notification responses.
type Processor interface {
	SetSendResult(ctx context.Context, res *SendResult) error

	Receive(ctx context.Context, callbackID string, result Result) error
	Start(context.Context, Dest) error
	Stop(context.Context, Dest) error
}
