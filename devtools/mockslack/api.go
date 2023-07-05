package mockslack

// API allows making calls to implemented Slack API methods.
//
// API methods implement permission/scope checking.
type API state

// API returns an API instance.
func (st *state) API() *API { return (*API)(st) }

// Channel represents a Slack channel or group.
type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	IsChannel  bool `json:"is_channel"`
	IsGroup    bool `json:"is_group"`
	IsArchived bool `json:"is_archived"`
}

// Message represents a Slack message.
type Message struct {
	TS        string `json:"ts"`
	ThreadTS  string `json:"thread_ts"`
	UpdateTS  string `json:"update_ts"`
	Text      string `json:"text"`
	User      string `json:"user"`
	Broadcast bool   `json:"reply_broadcast"`
	Color     string `json:"color"`

	ChannelID string
	ToUserID  string

	Actions []Action
}

type UserGroup struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Handle string `json:"handle"`

	IsUserGroup bool `json:"is_usergroup"`
}
