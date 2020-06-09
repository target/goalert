package alertlog

type EscalationMetaData struct {
	NewStepIndex    int
	Repeat          bool
	Forced          bool
	Deleted         bool
	OldDelayMinutes int
}

type NotificationMetaData struct {
	MessageID string
}

type CreatedMetaData struct {
	EPNoSteps bool
}
