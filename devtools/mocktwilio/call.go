package mocktwilio

import (
	"context"
)

type Call interface {
	ID() string

	From() string
	To() string

	// Text will return the last message returned by the application. It is empty until Answer is called.
	Text() string

	// User interactions below

	Answer(context.Context) error
	Hangup(context.Context, FinalCallStatus) error

	// Press will simulate a press of the specified key(s).
	//
	// It does nothing if the call isn't waiting for input.
	Press(context.Context, string) error

	// PressTimeout will simulate a user waiting for the menu to timeout.
	//
	// It does nothing if the call isn't waiting for input.
	PressTimeout(context.Context) error
}

type FinalCallStatus string

const (
	CallCompleted FinalCallStatus = "completed"
	CallFailed    FinalCallStatus = "failed"
	CallBusy      FinalCallStatus = "busy"
	CallNoAnswer  FinalCallStatus = "no-answer"
	CallCanceled  FinalCallStatus = "canceled"
)

type call struct {
	*callState
}

func (c *call) ID() string   { return c.callState.ID }
func (c *call) From() string { return c.callState.From }
func (c *call) To() string   { return c.callState.To }

func (srv *Server) Calls() <-chan Call { return srv.callCh }

// StartCall will start a new voice call.
func (srv *Server) StartCall(ctx context.Context, from, to string) error {
	return nil
}
