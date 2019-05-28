package mockslack

import (
	"context"
	"net/http"
	"strings"
)

// ConversationCreateOpts is used to configure a new
// channel or group.
type ConversationCreateOpts struct {
	Name     string
	Validate bool
}

// cleanChannelName will replace invalid characters with `_` and
// truncate the name if it is longer than 21 characters.
func cleanChannelName(name string) string {
	name = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			// make lower-case
			return r + ('a' - 'A')
		}
		if r >= '0' && r <= '9' {
			return r
		}
		if r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)

	if len(name) > 21 {
		name = name[:21]
	}

	return name
}

func validateChannelName(name string) error {
	if name == "" {
		return &response{Err: "invalid_name_required"}
	}
	if len(name) > 21 {
		return &response{Err: "invalid_name_maxlength"}
	}
	if name != cleanChannelName(name) {
		return &response{Err: "invalid_name"}
	}
	if !strings.ContainsAny(name, "abcdefghijklmnopqrstuvwxyz0123456789") {
		return &response{Err: "invalid_name_punctuation"}
	}

	return nil
}

// ChannelsCreate is used to create a channel.
func (st *API) ChannelsCreate(ctx context.Context, opts ConversationCreateOpts) (*Channel, error) {
	err := checkPermission(ctx, "channels:write")
	if err != nil {
		return nil, err
	}

	if !opts.Validate {
		opts.Name = cleanChannelName(opts.Name)
	}
	err = validateChannelName(opts.Name)
	if err != nil {
		return nil, err
	}

	ch := Channel{
		ID:        st.gen.ChannelID(),
		Name:      opts.Name,
		IsChannel: true,
	}

	st.mx.Lock()
	st.channels[ch.ID] = &channelState{Channel: ch}
	st.mx.Unlock()

	return &ch, nil
}

// ServeChannelsCreate serves a request to the `channels.create` API call.
//
// https://api.slack.com/methods/channels.create
func (s *Server) ServeChannelsCreate(w http.ResponseWriter, req *http.Request) {
	ch, err := s.API().ChannelsCreate(req.Context(), ConversationCreateOpts{Name: req.FormValue("name"), Validate: req.FormValue("validate") == "true"})
	if respondErr(w, err) {
		return
	}

	var resp struct {
		response
		Channel *Channel `json:"channel"`
	}
	resp.OK = true
	resp.Channel = ch

	respondWith(w, resp)
}
