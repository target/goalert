package notice

type Notice struct {
	Type    NoticeType `json:"type"`
	Message string     `json:"message"`
	Details string     `json:"details"`
}

// NoticeType represents the level of severity of a Notice.
type NoticeType int

const (
	Warning NoticeType = iota
	Error
	Info
)
