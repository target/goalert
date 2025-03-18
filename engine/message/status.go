package message

// Status represents the current state of an outgoing message.
type Status string

// Defined status values
const (
	// StatusPending means the message is waiting to be sent.
	StatusPending = Status("pending")

	// StatusSending means the message is in the process of being sent upstream
	StatusSending = Status("sending")

	// StatusQueuedRemotely means the message has been sent upstream, but is in a remote queue.
	StatusQueuedRemotely = Status("queued_remotely")

	// StatusSent means the message has been sent upstream, and has left the remote queue (if one exists).
	StatusSent = Status("sent")

	// StatusDelivered will be set on delivery if the upstream supports delivery confirmation.
	StatusDelivered = Status("delivered")

	// StatusRead will be set on read if the upstream supports read confirmation.
	StatusRead = Status("read")

	// StatusFailed means the message failed to send.
	StatusFailed = Status("failed")

	// StatusStale is used if the message expired before being sent.
	StatusStale = Status("stale")
)

// IsSent returns true if the message has been successfully sent to the downstream server.
func (s Status) IsSent() bool {
	switch s {
	case StatusQueuedRemotely, StatusDelivered, StatusSent, StatusRead:
		return true
	}
	return false
}
