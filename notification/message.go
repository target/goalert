package notification

//go:generate go run golang.org/x/tools/cmd/stringer -type MessageType

// A Message contains information that can be provided
// to a user for notification.
type Message interface {
	ID() string
	Type() MessageType
	Destination() Dest
	SubjectID() int
	Body() string
	ExtendedBody() string
}

// MessageType indicates the type of notification message.
type MessageType int

// Allowed types
const (
	MessageTypeAlert MessageType = iota
	MessageTypeAlertStatus
	MessageTypeTest
	MessageTypeVerification
)
