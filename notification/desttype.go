package notification

import (
	"crypto/sha256"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
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

type Dest struct {
	ID    DestID
	Type  DestType
	Value string
}

type DestHash [32]byte

// DestHash returns a comparable hash of the destination.
func (d Dest) DestHash() DestHash {
	return DestHash(sha256.Sum256([]byte(fmt.Sprintf("%s\n%s\n", d.Type.String(), d.Value))))
}

type SQLDest struct {
	CMID    uuid.NullUUID
	CMType  gadb.NullEnumUserContactMethodType
	CMValue sql.NullString

	NCID    uuid.NullUUID
	NCType  gadb.NullEnumNotifChannelType
	NCValue sql.NullString
}

func (s SQLDest) Dest() Dest {
	id := DestID{
		CMID: s.CMID,
		NCID: s.NCID,
	}
	if s.CMID.Valid {
		return Dest{
			ID:    id,
			Value: s.CMValue.String,
			Type:  ScannableDestType{CM: contactmethod.Type(s.CMType.EnumUserContactMethodType)}.DestType(),
		}
	}

	if s.NCID.Valid {
		return Dest{
			ID:    id,
			Value: s.NCValue.String,
			Type:  ScannableDestType{NC: notificationchannel.Type(s.NCType.EnumNotifChannelType)}.DestType(),
		}
	}

	panic("no valid ID")
}

// DestFromPair will return a Dest for a notification channel/contact method pair.
func DestFromPair(cm *contactmethod.ContactMethod, nc *notificationchannel.Channel) Dest {
	switch {
	case cm != nil:
		return Dest{
			ID: DestID{CMID: uuid.NullUUID{Valid: true, UUID: cm.ID}}, Value: cm.Value,
			Type: ScannableDestType{CM: cm.Type}.DestType(),
		}
	case nc != nil:
		return Dest{
			ID: DestID{NCID: uuid.NullUUID{Valid: true, UUID: nc.ID}}, Value: nc.Value,
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
	DestTypeChanWebhook
	DestTypeSlackUG
)

func (d Dest) String() string { return fmt.Sprintf("%s(%s)", d.Type.String(), d.ID) }

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
