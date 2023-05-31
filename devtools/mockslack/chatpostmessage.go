package mockslack

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ChatPostMessageOptions are parameters for a `chat.postMessage` call.
type ChatPostMessageOptions struct {
	ChannelID string
	Text      string
	Color     string

	Actions []Action

	AsUser bool

	User string

	UpdateTS  string
	ThreadTS  string
	Broadcast bool
}

func (ch *channelState) nextTS() string {
	t := time.Now()
	if !t.After(ch.TS) {
		t = ch.TS.Add(1)
	}
	ch.TS = t

	return strconv.FormatFloat(time.Duration(t.UnixNano()).Seconds(), 'f', -1, 64)
}

// ChatPostMessage posts a message to a channel.
func (st *API) ChatPostMessage(ctx context.Context, opts ChatPostMessageOptions) (*Message, error) {
	var err error
	var user string
	if opts.AsUser {
		err = checkPermission(ctx, "chat:write:user")
		user = userID(ctx)
	} else {
		err = checkPermission(ctx, "bot", "chat:write:bot")
		user = botID(ctx)
	}
	if err != nil {
		return nil, err
	}

	if len(opts.Text) > 40000 {
		return nil, &response{Err: "msg_too_long"}
	}

	st.mx.Lock()
	defer st.mx.Unlock()

	ch := st.channels[opts.ChannelID]
	if ch == nil && strings.HasPrefix(opts.ChannelID, "W") {
		// We need to create a new "channel" for the DM conversation.
		u, ok := st.users[opts.ChannelID]
		if !ok {
			return nil, &response{Err: "user_not_found"}
		}

		ch = &channelState{
			Channel: Channel{
				ID:        opts.ChannelID,
				Name:      "DM:" + u.Name,
				IsChannel: true,
			},
			Users: []string{userID(ctx)},
		}
		st.channels[opts.ChannelID] = ch
	}

	if ch == nil {
		if !st.flags.autoCreateChannel && !strings.HasPrefix(opts.ChannelID, "W") {
			return nil, &response{Err: "channel_not_found"}
		}

		// auto create channel
		ch = &channelState{Channel: Channel{
			ID:        opts.ChannelID,
			Name:      cleanChannelName(opts.ChannelID),
			IsChannel: true,
		}}
		if opts.AsUser {
			// add the user if needed
			ch.Users = append(ch.Users, userID(ctx))
		}

		st.channels[opts.ChannelID] = ch
	}

	if opts.AsUser && !contains(ch.Users, userID(ctx)) {
		return nil, &response{Err: "not_in_channel"}
	}

	if ch.IsArchived {
		return nil, &response{Err: "is_archived"}
	}

	msg := &Message{
		TS:    ch.nextTS(),
		Text:  opts.Text,
		User:  user,
		Color: opts.Color,

		ChannelID: opts.ChannelID,
		ToUserID:  opts.User,

		Actions: opts.Actions,

		UpdateTS: opts.UpdateTS,

		ThreadTS:  opts.ThreadTS,
		Broadcast: opts.Broadcast,
	}
	ch.Messages = append(ch.Messages, msg)

	return msg, nil
}

var errNoAttachment = errors.New("no attachment")

type attachments struct {
	Text    string
	Color   string
	Actions []Action
}
type Action struct {
	ChannelID string
	AppID     string
	TeamID    string

	BlockID  string
	ActionID string
	Text     string
	Value    string
	URL      string
}

// parseAttachments parses the attachments from the payload value.
func parseAttachments(appID, teamID, chanID, value string) (*attachments, error) {
	if value == "" {
		return nil, errNoAttachment
	}
	type textBlock struct{ Text string }

	var data [1]struct {
		Color  string
		Blocks []struct {
			Type     string
			BlockID  string `json:"block_id"`
			Elements json.RawMessage
			Text     textBlock
		}
	}
	err := json.Unmarshal([]byte(value), &data)
	if err != nil {
		return nil, err
	}

	var payload strings.Builder
	appendText := func(b textBlock) {
		if b.Text == "" {
			return
		}
		payload.WriteString(b.Text + "\n")
	}

	var actions []Action
	for _, b := range data[0].Blocks {
		appendText(b.Text)
		switch b.Type {
		case "context":
			var txtEl []textBlock
			err = json.Unmarshal(b.Elements, &txtEl)
			if err != nil {
				return nil, err
			}

			for _, e := range txtEl {
				appendText(e)
			}
		case "actions":
			var acts []struct {
				Text     textBlock
				ActionID string `json:"action_id"`
				Value    string
				URL      string
			}
			err = json.Unmarshal(b.Elements, &acts)
			if err != nil {
				return nil, err
			}

			for _, a := range acts {
				actions = append(actions, Action{
					ChannelID: chanID,
					TeamID:    teamID,
					AppID:     appID,
					BlockID:   b.BlockID,
					ActionID:  a.ActionID,
					Text:      a.Text.Text,
					Value:     a.Value,
					URL:       a.URL,
				})
			}
		default:
			continue
		}

	}

	return &attachments{
		Text:    payload.String(),
		Color:   data[0].Color,
		Actions: actions,
	}, nil
}

// ServeChatPostMessage serves a request to the `chat.postMessage` API call.
//
// https://api.slack.com/methods/chat.postMessage
func (s *Server) ServeChatPostMessage(w http.ResponseWriter, req *http.Request) {
	s.serveChatPostMessage(w, req, false)
}

// ServeChatUpdate serves a request to the `chat.update` API call.
//
// https://api.slack.com/methods/chat.update
func (s *Server) ServeChatUpdate(w http.ResponseWriter, req *http.Request) {
	s.serveChatPostMessage(w, req, true)
}

func (s *Server) serveChatPostMessage(w http.ResponseWriter, req *http.Request, isUpdate bool) {
	chanID := req.FormValue("channel")

	var text, color string
	var actions []Action

	var appID string
	s.mx.Lock()
	for id := range s.apps {
		if appID != "" {
			panic("multiple apps not supported")
		}
		appID = id
	}
	s.mx.Unlock()

	attachment, err := parseAttachments(appID, s.teamID, chanID, req.FormValue("attachments"))
	if err == errNoAttachment {
		err = nil
		text = req.FormValue("text")
	} else {
		text = attachment.Text
		color = attachment.Color
		actions = attachment.Actions
	}
	if respondErr(w, err) {
		return
	}

	var updateTS string
	if isUpdate {
		updateTS = req.FormValue("ts")
	}

	msg, err := s.API().ChatPostMessage(req.Context(), ChatPostMessageOptions{
		ChannelID: chanID,
		Text:      text,
		Color:     color,
		Actions:   actions,
		AsUser:    req.FormValue("as_user") == "true",
		ThreadTS:  req.FormValue("thread_ts"),
		UpdateTS:  updateTS,

		User: req.FormValue("user"),

		Broadcast: req.FormValue("reply_broadcast") == "true",
	})
	if respondErr(w, err) {
		return
	}

	var respData struct {
		response
		TS      string
		Channel string   `json:"channel"`
		Message *Message `json:"message"`
	}
	respData.TS = msg.TS
	respData.OK = true
	respData.Channel = chanID
	respData.Message = msg

	respondWith(w, respData)
}
