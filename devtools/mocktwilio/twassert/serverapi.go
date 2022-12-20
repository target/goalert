package twassert

import (
	"context"

	"github.com/target/goalert/devtools/mocktwilio"
)

// ServerAPI is the interface for the mocktwilio server.
type ServerAPI interface {
	SendMessage(ctx context.Context, from, to, body string) (mocktwilio.Message, error)

	// WaitInFlight should return after all in-flight messages are processed.
	WaitInFlight(context.Context) error

	Messages() <-chan mocktwilio.Message
	Calls() <-chan mocktwilio.Call
}
