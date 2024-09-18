package nfymsg

// User provides information about a user for notifications.
type User struct {
	ID   string
	Name string
	URL  string
}

// ScheduleOnCallUsers is a Message that indicates which users are
// currently on-call for a Schedule
type ScheduleOnCallUsers struct {
	Base

	ScheduleID   string
	ScheduleName string
	ScheduleURL  string

	Users []User
}
