package nfymsg

import "github.com/target/goalert/gadb"

// A Message contains information that can be provided
// to a user for notification.
type Message interface {
	ID() string
	Dest() gadb.DestV1
	DestArg(name string) string
}

type Base struct {
	MsgID   string
	MsgDest gadb.DestV1
}

func (b Base) ID() string        { return b.MsgID }
func (b Base) Dest() gadb.DestV1 { return b.MsgDest }
func (b Base) DestArg(name string) string {
	if b.MsgDest.Args == nil {
		return ""
	}
	return b.MsgDest.Args[name]
}
