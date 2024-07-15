package slack

const (
	DestTypeSlackDirectMessage = "builtin-slack-dm"
	DestTypeSlackChannel       = "builtin-slack-channel"
	DestTypeSlackUsergroup     = "builtin-slack-usergroup"

	FieldSlackUserID      = "slack_user_id"
	FieldSlackChannelID   = "slack_channel_id"
	FieldSlackUsergroupID = "slack_usergroup_id"

	FallbackIconURL = "builtin://slack"
)
