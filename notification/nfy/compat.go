package nfy

// DestType definitions for compatibility (should be moved to individual packages).
const (
	CompatDestTypeTwilioSMS   = "builtin-twilio-sms"
	CompatDestTypeTwilioVoice = "builtin-twilio-voice"
	CompatDestTypeSMTP        = "builtin-smtp-email"
	CompatDestTypeWebhook     = "builtin-webhook"
	CompatDestTypeSlackDM     = "builtin-slack-dm"
	CompatDestTypeSlackChan   = "builtin-slack-channel"
	CompatDestTypeSlackUG     = "builtin-slack-usergroup"
)
