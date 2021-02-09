package notification

//go:generate go run golang.org/x/tools/cmd/stringer -type DestType

type Dest struct {
	ID    string
	Type  DestType
	Value string
}

type DestType int

const (
	DestTypeUnknown DestType = iota
	DestTypeVoice
	DestTypeSMS
	DestTypeSlackChannel
	DestTypeUserEmail
	DestTypeUserWebhook
)

// IsUserCM returns true if the DestType represents a user contact method.
func (t DestType) IsUserCM() bool {
	switch t {
	case DestTypeSMS, DestTypeVoice, DestTypeUserEmail, DestTypeUserWebhook:
		return true
	}
	return false
}
