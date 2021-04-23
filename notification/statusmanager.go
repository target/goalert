package notification

import "context"

type namedReceiver struct {
	Receiver
	ns *namedSender
}

// UpdateStatus calls the underlying UpdateStatus method after wrapping the status for the
// namedSender.
func (nr *namedReceiver) UpdateStatus(ctx context.Context, status *MessageStatus) error {
	return nr.Receiver.UpdateStatus(ctx, status.wrap(ctx, nr.ns))
}
