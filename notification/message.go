package notification

// A Message contains information that can be provided
// to a user for notification.
type Message interface {
	ID() string
	Type() MessageType
	Destination() Dest
}
