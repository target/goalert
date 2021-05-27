package notification

// ScheduleOnCallStatus represents notification of an on-call assignment

type Schedule struct {
	ID   string
	Name string
	URL  string
}

type User struct {
	ID   string
	Name string
	URL  string
}
type ScheduleOnCallStatus struct {
	Dest       Dest
	CallbackID string

	Schedule
	Users []User
}

var _ Message = &ScheduleOnCallStatus{}

func (s ScheduleOnCallStatus) ID() string        { return s.CallbackID }
func (s ScheduleOnCallStatus) Destination() Dest { return s.Dest }
func (s ScheduleOnCallStatus) Type() MessageType { return MessageTypeScheduleOnCallStatus }
