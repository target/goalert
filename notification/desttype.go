package notification

import (
	"github.com/google/uuid"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/user/contactmethod"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type DestType

type DestID struct {
	// CMID is the ID of the user contact method.
	CMID uuid.NullUUID
	// NCID is the ID of the notification channel.
	NCID uuid.NullUUID
}

func (d DestID) IsUserCM() bool { return d.CMID.Valid }
func (d DestID) String() string {
	switch {
	case d.CMID.Valid:
		return d.CMID.UUID.String()
	case d.NCID.Valid:
		return d.NCID.UUID.String()
	}
	return ""
}

func (d DestID) UUID() uuid.UUID {
	switch {
	case d.CMID.Valid:
		return d.CMID.UUID
	case d.NCID.Valid:
		return d.NCID.UUID
	}
	return uuid.Nil
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
	DestTypeChanWebhook
	DestTypeSlackUG
)

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
	case notificationchannel.TypeSlackChan:
		return DestTypeSlackChannel
	case notificationchannel.TypeWebhook:
		return DestTypeChanWebhook
	case notificationchannel.TypeSlackUG:
		return DestTypeSlackUG
	}

	return DestTypeUnknown
}

// NCType returns the notificationchannel.Type associated with the DestType.
func (t DestType) NCType() notificationchannel.Type {
	switch t {
	case DestTypeSlackChannel:
		return notificationchannel.TypeSlackChan
	case DestTypeChanWebhook:
		return notificationchannel.TypeWebhook
	case DestTypeSlackUG:
		return notificationchannel.TypeSlackUG
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
