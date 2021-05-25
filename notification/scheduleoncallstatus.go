package notification

// ScheduleOnCallStatus represents notification of an on-call assignment
type ScheduleOnCallStatus struct {
	Dest       Dest
	CallbackID string

	Schedule struct {
		ID   string
		Name string
		URL  string
	}
	Users []struct {
		ID   string
		Name string
		URL  string
	}
}

var _ Message = &ScheduleOnCallStatus{}

func (s ScheduleOnCallStatus) ID() string        { return s.CallbackID }
func (s ScheduleOnCallStatus) Destination() Dest { return s.Dest }
func (s ScheduleOnCallStatus) Type() MessageType { return MessageTypeScheduleOnCallStatus }
