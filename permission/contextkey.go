package permission

type contextKey int

const (
	contextKeyUserRole contextKey = iota
	contextKeyUserID
	contextKeyCheckCount
	contextKeyServiceID
	contextKeySystem
	contextHasAuth
	contextKeyTeamID
	contextKeyCheckCountMax
	contextKeySourceInfo
)
