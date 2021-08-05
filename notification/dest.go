package notification

import (
	"fmt"

	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/user/contactmethod"
)

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

func (d Dest) String() string { return fmt.Sprintf("%s(%s)", d.Type.String(), d.ID) }

// IsUserCM returns true if the DestType represents a user contact method.
func (t DestType) IsUserCM() bool { return t.CMType() != contactmethod.TypeUnknown }

// ScannableDestType allows scanning a DestType from separate columns for user contact methods and notification channels.
type ScannableDestType struct {
	CM contactmethod.Type
	NC notificationchannel.Type
}

// DestType returns a DestType from the scanned values.
func (t ScannableDestType) DestType() DestType { return coalesceDestType(t.CM, t.NC) }

func destTypeFromCM(t contactmethod.Type) DestType {
	switch t {
	case contactmethod.TypeSMS:
		return DestTypeSMS
	case contactmethod.TypeVoice:
		return DestTypeVoice
	case contactmethod.TypeEmail:
		return DestTypeUserEmail
	case contactmethod.TypeWebhook:
		return DestTypeUserWebhook
	}

	return DestTypeUnknown
}

func destTypeFromNC(t notificationchannel.Type) DestType {
	switch t {
	case notificationchannel.TypeSlack:
		return DestTypeSlackChannel
	}

	return DestTypeUnknown
}

func coalesceDestType(cm contactmethod.Type, nc notificationchannel.Type) DestType {
	if nc != notificationchannel.TypeUnknown {
		return destTypeFromNC(nc)
	}

	return destTypeFromCM(cm)
}

// NCType returns the notificationchannel.Type associated with the DestType.
func (t DestType) NCType() notificationchannel.Type {
	switch t {
	case DestTypeSlackChannel:
		return notificationchannel.TypeSlack
	}

	return notificationchannel.TypeUnknown
}

// CMType returns the contactmethod.Type associated with the DestType.
func (t DestType) CMType() contactmethod.Type {
	switch t {
	case DestTypeSMS:
		return contactmethod.TypeSMS
	case DestTypeVoice:
		return contactmethod.TypeVoice
	case DestTypeUserEmail:
		return contactmethod.TypeEmail
	case DestTypeUserWebhook:
		return contactmethod.TypeWebhook
	}

	return contactmethod.TypeUnknown
}
