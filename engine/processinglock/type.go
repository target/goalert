package processinglock

// Type indicates the lock type. For TypeMessage, the RegionID is used.
type Type string

// Recognized types
const (
	TypeEscalation   Type = "escalation"
	TypeHeartbeat    Type = "heartbeat"
	TypeNPCycle      Type = "np_cycle"
	TypeRotation     Type = "rotation"
	TypeSchedule     Type = "schedule"
	TypeStatusUpdate Type = "status_update"
	TypeVerify       Type = "verify"
	TypeMessage      Type = "message"
	TypeCleanup      Type = "cleanup"
	TypeMetrics      Type = "metrics"
	TypeCompat       Type = "compat"
)
