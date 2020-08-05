package notice

type Notice struct {
	Type    NoticeType
	Message string
	Details string
}

// NoticeType represents the level of severity of a Notice.
type NoticeType int

// Defaults to Warning when unset
const (
	Warning NoticeType = iota
	Error
	Info
)
