package notification

import "context"

type namedReceiver struct {
	r  ResultReceiver
	ns *namedSender
}

var _ Receiver = &namedReceiver{}

// UpdateStatus calls the underlying UpdateStatus method after wrapping the status for the
// namedSender.
func (nr *namedReceiver) SetMessageStatus(ctx context.Context, externalID string, status *Status) error {
	res := &SendResult{Status: *status}
	res.ProviderMessageID.Provider = nr.ns.name
	res.ProviderMessageID.ID = externalID
	return nr.r.SetSendResult(ctx, res)
}

// Start implements the Receiver interface by calling the underlying Receiver.Start method.
func (nr *namedReceiver) Start(ctx context.Context, d Dest) error {
	metricRecvTotal.WithLabelValues(d.Type.String(), "START")
	return nr.r.Start(ctx, d)
}

// Stop implements the Receiver interface by calling the underlying Receiver.Stop method.
func (nr *namedReceiver) Stop(ctx context.Context, d Dest) error {
	metricRecvTotal.WithLabelValues(d.Type.String(), "STOP")
	return nr.r.Stop(ctx, d)
}

// Receive implements the Receiver interface by calling the underlying Receiver.Receive method.
func (nr *namedReceiver) Receive(ctx context.Context, callbackID string, result Result) error {
	metricRecvTotal.WithLabelValues(nr.ns.destType.String(), result.String())
	return nr.r.Receive(ctx, callbackID, result)
}
