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

// DestFromPair will return a Dest for a notification channel/contact method pair.
func DestFromPair(cm *contactmethod.ContactMethod, nc *notificationchannel.Channel) Dest {
	switch {
	case cm != nil:
		return Dest{
			ID: cm.ID, Value: cm.Value,
			Type: ScannableDestType{CM: cm.Type}.DestType(),
		}
	case nc != nil:
		return Dest{
			ID: nc.ID, Value: nc.Value,
			Type: ScannableDestType{NC: nc.Type}.DestType(),
		}
	}
	return Dest{}
}

// DestType represents the type of destination, it is a combination of available contact methods and notification channels.
type DestType int

const (
	DestTypeUnknown DestType = iota
	DestTypeVoice
	DestTypeSMS
	DestTypeSlackChannel
	DestTypeSlackDM
	DestTypeUserEmail
	DestTypeUserWebhook
)

func (d Dest) String() string { return fmt.Sprintf("%s(%s)", d.Type.String(), d.ID) }

// IsUserCM returns true if the DestType represents a user contact method.
func (t DestType) IsUserCM() bool { return t.CMType() != contactmethod.TypeUnknown }

// ScannableDestType allows scanning a DestType from separate columns for user contact methods and notification channels.
type ScannableDestType struct {
	// CM is the contactmethod.Type and should be scanned from the `type` column from `user_contact_methods`.
	CM contactmethod.Type

	// NC is the notificationchannel.Type and should be scanned from the `type` column from `notification_channels`.
	NC notificationchannel.Type
}

// DestType returns a DestType from the scanned values.
func (t ScannableDestType) DestType() DestType {
	switch t.CM {
	case contactmethod.TypeSMS:
		return DestTypeSMS
	case contactmethod.TypeVoice:
		return DestTypeVoice
	case contactmethod.TypeEmail:
		return DestTypeUserEmail
	case contactmethod.TypeWebhook:
		return DestTypeUserWebhook
	case contactmethod.TypeSlackDM:
		return DestTypeSlackDM
	}

	switch t.NC {
	case notificationchannel.TypeSlack:
		return DestTypeSlackChannel
	}

	return DestTypeUnknown
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
	case DestTypeSlackDM:
		return contactmethod.TypeSlackDM
	}

	return contactmethod.TypeUnknown
}
