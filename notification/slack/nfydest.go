package slack

import "github.com/target/goalert/gadb"

const (
	DestTypeSlackDirectMessage = "builtin-slack-dm"
	DestTypeSlackChannel       = "builtin-slack-channel"
	DestTypeSlackUsergroup     = "builtin-slack-usergroup"

	FieldSlackUserID      = "slack_user_id"
	FieldSlackChannelID   = "slack_channel_id"
	FieldSlackUsergroupID = "slack_usergroup_id"

	FallbackIconURL = "builtin://slack"
)

func NewDirectMessageDest(slackUserID string) gadb.DestV1 {
	return gadb.NewDestV1(DestTypeSlackDirectMessage, FieldSlackUserID, slackUserID)
}

func NewChannelDest(slackChannelID string) gadb.DestV1 {
	return gadb.NewDestV1(DestTypeSlackChannel, FieldSlackChannelID, slackChannelID)
}

func NewUsergroupDest(slackUsergroupID, errChanID string) gadb.DestV1 {
	return gadb.NewDestV1(DestTypeSlackUsergroup,
		FieldSlackUsergroupID, slackUsergroupID,
		FieldSlackChannelID, errChanID)
}
