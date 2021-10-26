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

	AsUser bool

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
	if ch == nil {
		if !st.flags.autoCreateChannel {
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

		ThreadTS:  opts.ThreadTS,
		Broadcast: opts.Broadcast,
	}
	ch.Messages = append(ch.Messages, msg)

	return msg, nil
}

var errNoAttachment = errors.New("no attachment")

func attachmentsText(value string) (text, color string, err error) {
	if value == "" {
		return "", "", errNoAttachment
	}

	type textBlock struct{ Text string }

	var data [1]struct {
		Color  string
		Blocks []struct {
			Elements []textBlock
			Text     textBlock
		}
	}

	err = json.Unmarshal([]byte(value), &data)
	if err != nil {
		return "", "", err
	}

	var payload strings.Builder
	appendText := func(b textBlock) {
		if b.Text == "" {
			return
		}
		payload.WriteString(b.Text + "\n")
	}

	for _, b := range data[0].Blocks {
		appendText(b.Text)
		for _, e := range b.Elements {
			appendText(e)
		}
	}

	return payload.String(), data[0].Color, nil
}

// ServeChatPostMessage serves a request to the `chat.postMessage` API call.
//
// https://api.slack.com/methods/chat.postMessage
func (s *Server) ServeChatPostMessage(w http.ResponseWriter, req *http.Request) {
	chanID := req.FormValue("channel")

	text, color, err := attachmentsText(req.FormValue("attachments"))
	if err == errNoAttachment {
		err = nil
		text = req.FormValue("text")
	}
	if respondErr(w, err) {
		return
	}
	msg, err := s.API().ChatPostMessage(req.Context(), ChatPostMessageOptions{
		ChannelID: chanID,
		Text:      text,
		Color:     color,
		AsUser:    req.FormValue("as_user") == "true",
		ThreadTS:  req.FormValue("thread_ts"),

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
