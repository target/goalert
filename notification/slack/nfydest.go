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

func NewChannelDest(id string) gadb.DestV1 {
	return gadb.NewDestV1(DestTypeSlackChannel, FieldSlackChannelID, id)
}

func NewDirectMessageDest(id string) gadb.DestV1 {
	return gadb.NewDestV1(DestTypeSlackDirectMessage, FieldSlackUserID, id)
}

func NewUsergroupDest(groupID, channelID string) gadb.DestV1 {
	return gadb.NewDestV1(DestTypeSlackUsergroup, FieldSlackUsergroupID, groupID, FieldSlackChannelID, channelID)
}
