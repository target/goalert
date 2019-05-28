package mockslack

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// ChatPostMessageOptions are parameters for a `chat.postMessage` call.
type ChatPostMessageOptions struct {
	ChannelID string
	Text      string

	AsUser bool

	ThreadTS string
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
		TS:   ch.nextTS(),
		Text: opts.Text,
		User: user,
	}
	ch.Messages = append(ch.Messages, msg)

	return msg, nil
}

// ServeChatPostMessage serves a request to the `chat.postMessage` API call.
//
// https://api.slack.com/methods/chat.postMessage
func (s *Server) ServeChatPostMessage(w http.ResponseWriter, req *http.Request) {
	chanID := req.FormValue("channel")
	msg, err := s.API().ChatPostMessage(req.Context(), ChatPostMessageOptions{
		ChannelID: chanID,
		Text:      req.FormValue("text"),
		AsUser:    req.FormValue("as_user") == "true",
		ThreadTS:  req.FormValue("thread_ts"),
	})
	if respondErr(w, err) {
		return
	}

	var respData struct {
		response
		Channel string   `json:"channel"`
		Message *Message `json:"message"`
	}
	respData.OK = true
	respData.Channel = chanID
	respData.Message = msg

	respondWith(w, respData)
}
