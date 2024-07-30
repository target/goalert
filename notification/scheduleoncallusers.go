package notification

import "github.com/target/goalert/gadb"

// User provides information about a user for notifications.
type User struct {
	ID   string
	Name string
	URL  string
}

// ScheduleOnCallUsers is a Message that indicates which users are
// currently on-call for a Schedule
type ScheduleOnCallUsers struct {
	Dest       gadb.DestV1
	CallbackID string

	ScheduleID   string
	ScheduleName string
	ScheduleURL  string

	Users []User
}

var _ Message = &ScheduleOnCallUsers{}

func (s ScheduleOnCallUsers) ID() string               { return s.CallbackID }
func (s ScheduleOnCallUsers) Destination() gadb.DestV1 { return s.Dest }
func (s ScheduleOnCallUsers) Type() MessageType        { return MessageTypeScheduleOnCallUsers }
