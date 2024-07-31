package notification

import (
	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
)

type ProviderMessageID = gadb.ProviderMessageID

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
