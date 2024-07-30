package notification

import "github.com/target/goalert/gadb"

// Test represents outgoing test notification.
type Test struct {
	Dest       gadb.DestV1
	CallbackID string // CallbackID is the identifier used to communicate a response to the notification
}

var _ Message = &Test{}

func (t Test) Type() MessageType        { return MessageTypeTest }
func (t Test) ID() string               { return t.CallbackID }
func (t Test) Destination() gadb.DestV1 { return t.Dest }
func (t Test) Body() string             { return "" }
func (t Test) ExtendedBody() string     { return "" }
func (t Test) SubjectID() int           { return -1 }
