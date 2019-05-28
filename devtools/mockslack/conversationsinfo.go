package mockslack

import (
	"context"
	"net/http"
)

// ConversationsInfo returns information about a conversation.
func (st *API) ConversationsInfo(ctx context.Context, id string) (*Channel, error) {
	err := checkPermission(ctx, "bot", "channels:read", "groups:read", "im:read", "mpim:read")
	if err != nil {
		return nil, err
	}

	st.mx.Lock()
	defer st.mx.Unlock()

	ch := st.channels[id]
	if ch == nil {
		return nil, &response{Err: "channel_not_found"}
	}

	if hasScope(ctx, "bot") {
		return &ch.Channel, nil
	}

	if ch.IsGroup {
		err = checkPermission(ctx, "groups:read")
	} else {
		err = checkPermission(ctx, "channels:read")
	}
	if err != nil {
		return nil, err
	}

	if ch.IsGroup && !contains(ch.Users, userID(ctx)) {
		// user is not a member of the group
		return nil, &response{Err: "channel_not_found"}
	}

	return &ch.Channel, nil
}

// ServeConversationsInfo serves a request to the `conversations.info` API call.
//
// https://api.slack.com/methods/conversations.info
func (s *Server) ServeConversationsInfo(w http.ResponseWriter, req *http.Request) {
	ch, err := s.API().ConversationsInfo(req.Context(), req.FormValue("channel"))
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
