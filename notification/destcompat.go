package notification

import (
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
