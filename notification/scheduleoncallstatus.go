package notification

// User provides information about a user for notifications.
type User struct {
	ID   string
	Name string
	URL  string
}

// ScheduleOnCallStatus is a Message that indicates which users are
// currently on-call for a Schedule
type ScheduleOnCallStatus struct {
	Dest       Dest
	CallbackID string

	ScheduleID   string
	ScheduleName string
	ScheduleURL  string

	Users []User
}

var _ Message = &ScheduleOnCallStatus{}

func (s ScheduleOnCallStatus) ID() string        { return s.CallbackID }
func (s ScheduleOnCallStatus) Destination() Dest { return s.Dest }
func (s ScheduleOnCallStatus) Type() MessageType { return MessageTypeScheduleOnCallNotification }
