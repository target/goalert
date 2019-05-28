package alertlog

type EscalationMetaData struct {
	NewStepIndex    int
	Repeat          bool
	Forced          bool
	Deleted         bool
	OldDelayMinutes int
}

type NotificationMetaData struct {
	UserID string
	CMType string
}
