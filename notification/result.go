package notification

//go:generate go run golang.org/x/tools/cmd/stringer -type Result

// Result specifies a response to a notification.
type Result int

// Possible notification responses.
const (
	ResultAcknowledge Result = iota
	ResultResolve
	ResultEscalate
)
