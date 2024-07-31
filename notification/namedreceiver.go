package notification

import (
	"context"

	"github.com/target/goalert/auth/authlink"
	"github.com/target/goalert/gadb"
)

type namedReceiver struct {
	r  ResultReceiver
	ns *namedSender
}

var _ Receiver = &namedReceiver{}

// IsKnownDest calls the underlying ResultReceiver.IsKnownDest method for the current type.
func (nr *namedReceiver) IsKnownDest(ctx context.Context, dest gadb.DestV1) (bool, error) {
	return nr.r.IsKnownDest(ctx, dest)
}

// SetMessageStatus calls the underlying ResultReceiver's SetSendResult method after wrapping the status for the
// namedSender.
func (nr *namedReceiver) SetMessageStatus(ctx context.Context, externalID string, status *Status) error {
	res := &SendResult{Status: *status}
	res.ProviderMessageID.ProviderName = nr.ns.name
	res.ProviderMessageID.ExternalID = externalID
	return nr.r.SetSendResult(ctx, res)
}

// AuthLinkURL calls the underlying AuthLinkURL method.
func (nr *namedReceiver) AuthLinkURL(ctx context.Context, providerID, subjectID string, meta authlink.Metadata) (string, error) {
	return nr.r.AuthLinkURL(ctx, providerID, subjectID, meta)
}

// Start implements the Receiver interface by calling the underlying Receiver.Start method.
func (nr *namedReceiver) Start(ctx context.Context, d gadb.DestV1) error {
	metricRecvTotal.WithLabelValues(d.Type, "START")
	return nr.r.Start(ctx, d)
}

// Stop implements the Receiver interface by calling the underlying Receiver.Stop method.
func (nr *namedReceiver) Stop(ctx context.Context, d gadb.DestV1) error {
	metricRecvTotal.WithLabelValues(d.Type, "STOP")
	return nr.r.Stop(ctx, d)
}

// Receive implements the Receiver interface by calling the underlying Receiver.Receive method.
func (nr *namedReceiver) Receive(ctx context.Context, callbackID string, result Result) error {
	metricRecvTotal.WithLabelValues(nr.ns.destType, result.String())
	return nr.r.Receive(ctx, callbackID, result)
}

// Receive implements the Receiver interface by calling the underlying Receiver.ReceiveSubject method.
func (nr *namedReceiver) ReceiveSubject(ctx context.Context, providerID, subjectID, callbackID string, result Result) error {
	metricRecvTotal.WithLabelValues(nr.ns.destType, result.String())
	return nr.r.ReceiveSubject(ctx, providerID, subjectID, callbackID, result)
}
