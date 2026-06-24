package nfymsg

// AlertBundle represents a bundle of outgoing alert notifications for a single service.
type AlertBundle struct {
	Base

	ServiceID   string
	ServiceName string // The service being notified for
	Count       int    // Number of unacked alerts
}
