package notification

import (
	"github.com/target/goalert/notification/nfy"
)

type Dest = nfy.Dest

// DestType represents the type of destination, it is a combination of available contact methods and notification channels.
type DestType = nfy.DestType

const (
	DestTypeUnknown      = ""
	DestTypeVoice        = "builtin-twilio-voice"
	DestTypeSMS          = "builtin-twilio-sms"
	DestTypeSlackChannel = "builtin-slack-channel"
	DestTypeSlackDM      = "builtin-slack-dm"
	DestTypeUserEmail    = "builtin-smtp-email"
	DestTypeWebhook      = "builtin-webhook"
	DestTypeSlackUG      = "builtin-slack-usergroup"
)
