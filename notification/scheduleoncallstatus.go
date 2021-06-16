package notification

// Schedule provides information about a schedule for notifications.
type Schedule struct {
	ID   string
	Name string
	URL  string
}

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

	Schedule Schedule
	Users    []User
}

var _ Message = &ScheduleOnCallStatus{}

func (s ScheduleOnCallStatus) ID() string        { return s.CallbackID }
func (s ScheduleOnCallStatus) Destination() Dest { return s.Dest }
func (s ScheduleOnCallStatus) Type() MessageType { return MessageTypeScheduleOnCallStatus }
