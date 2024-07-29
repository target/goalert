package notification

import (
	"strings"

	"github.com/target/goalert/gadb"
)

type ProviderMessageID = gadb.ProviderMessageID

func DestV1TypeToDestType(t string) DestType {
	switch t {
	case "builtin-twilio-voice":
		return DestTypeVoice
	case "builtin-twilio-sms":
		return DestTypeSMS
	case "builtin-slack-channel":
		return DestTypeSlackChannel
	case "builtin-slack-dm":
		return DestTypeSlackDM
	case "builtin-smtp-email":
		return DestTypeUserEmail
	case "builtin-webhook":
		return DestTypeUserWebhook
	case "builtin-slack-usergroup":
		return DestTypeSlackUG
	}
	return DestTypeUnknown
}

// ToDestV1 converts a Dest to a DestV1.
func (d Dest) ToDestV1() gadb.DestV1 {
	// The values here are copied from the graphql package until we have a centralized location for these values.
	//
	// This method is also subject to deletion once Type/Value are no longer in use.

	switch d.Type {
	case DestTypeVoice:
		return gadb.DestV1{
			Type: "builtin-twilio-voice",
			Args: map[string]string{"phone_number": d.Value},
		}
	case DestTypeSMS:
		return gadb.DestV1{
			Type: "builtin-twilio-sms",
			Args: map[string]string{"phone_number": d.Value},
		}
	case DestTypeSlackChannel:
		return gadb.DestV1{
			Type: "builtin-slack-channel",
			Args: map[string]string{"slack_channel_id": d.Value},
		}
	case DestTypeSlackDM:
		return gadb.DestV1{
			Type: "builtin-slack-dm",
			Args: map[string]string{"slack_user_id": d.Value},
		}
	case DestTypeUserEmail:
		return gadb.DestV1{
			Type: "builtin-smtp-email",
			Args: map[string]string{"email_address": d.Value},
		}
	case DestTypeUserWebhook, DestTypeChanWebhook:
		return gadb.DestV1{
			Type: "builtin-webhook",
			Args: map[string]string{"webhook_url": d.Value},
		}
	case DestTypeSlackUG:
		ugID, chanID, _ := strings.Cut(d.Value, ":")
		return gadb.DestV1{
			Type: "builtin-slack-usergroup",
			Args: map[string]string{
				"slack_usergroup_id": ugID,
				"slack_channel_id":   chanID,
			},
		}
	}

	panic("unsupported destination type " + d.Type.String())
}
