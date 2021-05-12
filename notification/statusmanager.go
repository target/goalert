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

// Start implements the Receiver interface by calling the underlying Receiver.Start method.
func (nr *namedReceiver) Start(ctx context.Context, d Dest) error {
	metricRecvTotal.WithLabelValues(d.Type.String(), "START")
	return nr.Receiver.Start(ctx, d)
}

// Stop implements the Receiver interface by calling the underlying Receiver.Stop method.
func (nr *namedReceiver) Stop(ctx context.Context, d Dest) error {
	metricRecvTotal.WithLabelValues(d.Type.String(), "STOP")
	return nr.Receiver.Stop(ctx, d)
}

// Receive implements the Receiver interface by calling the underlying Receiver.Receive method.
func (nr *namedReceiver) Receive(ctx context.Context, callbackID string, result Result) error {
	metricRecvTotal.WithLabelValues(nr.ns.destType.String(), result.String())
	return nr.Receiver.Receive(ctx, callbackID, result)
}
