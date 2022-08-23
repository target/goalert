package mocktwilio

import "context"

type Call interface {
	ID() string

	From() string
	To() string

	// Text will return the last message returned by the application. It is empty until Answer is called.
	Text() string

	Answer(context.Context) error

	IsActive() bool

	// Press will simulate a press of the specified key.
	//
	// It does nothing if Answer has not been called or
	// the call has ended.
	Press(context.Context, string) error

	End(context.Context, FinalCallStatus) error
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
