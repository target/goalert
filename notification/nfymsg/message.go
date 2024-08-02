package nfymsg

import "github.com/target/goalert/gadb"

// A Message contains information that can be provided
// to a user for notification.
type Message interface {
	MsgID() string
	DestType() string
	DestArg(name string) string
}

type Base struct {
	ID   string
	Dest gadb.DestV1
}

func (b Base) Base() Base       { return b }
func (b Base) MsgID() string    { return b.ID }
func (b Base) DestType() string { return b.Dest.Type }
func (b Base) DestArg(name string) string {
	if b.Dest.Args == nil {
		return ""
	}
	return b.Dest.Args[name]
}
