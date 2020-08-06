package notice

type Notice struct {
	Type    Type
	Message string
	Details string
}

// NoticeType represents the level of severity of a Notice.
type Type int

// Defaults to Warning when unset
const (
	TypeWarning Type = iota
	TypeError
	TypeInfo
)
