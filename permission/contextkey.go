package permission

type contextKey int

const (
	contextKeyUserRole contextKey = iota
	contextKeyUserID
	contextKeyServiceID
	contextKeySystem
	contextHasAuth
	contextKeyTeamID
	contextKeySourceInfo
)
