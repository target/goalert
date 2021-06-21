package notification

// User provides information about a user for notifications.
type User struct {
	ID   string
	Name string
	URL  string
}

// ScheduleOnCallNotification is a Message that indicates which users are
// currently on-call for a Schedule
type ScheduleOnCallNotification struct {
	Dest       Dest
	CallbackID string

	ScheduleID   string
	ScheduleName string
	ScheduleURL  string

	Users []User
}

var _ Message = &ScheduleOnCallNotification{}

func (s ScheduleOnCallNotification) ID() string        { return s.CallbackID }
func (s ScheduleOnCallNotification) Destination() Dest { return s.Dest }
func (s ScheduleOnCallNotification) Type() MessageType { return MessageTypeScheduleOnCallNotification }
