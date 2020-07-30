package notice

type Notice struct {
	Type    NoticeType `json:"type"`
	Message string     `json:"message"`
	Details string     `json:"details"`
}

// MessageState represents the current state of an outgoing message.
type NoticeType int

const (
	Warning NoticeType = iota
	Error
	Info
)
