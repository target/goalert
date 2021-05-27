package notification

import "fmt"

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
)

func (d Dest) String() string { return fmt.Sprintf("%s(%s)", d.Type.String(), d.ID) }

// IsUserCM returns true if the DestType represents a user contact method.
func (t DestType) IsUserCM() bool {
	switch t {
	case DestTypeSMS, DestTypeVoice, DestTypeUserEmail:
		return true
	}
	return false
}
