package mockslack

import (
	"context"
	"encoding/base64"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// ConversationsListOpts contains parameters for the ConversationsList API call.
type ConversationsListOpts struct {
	Cursor          string
	ExcludeArchived bool
	Limit           int
	Types           string
}

// ConversationsList returns a list of channel-like conversations in a workspace.
func (st *API) ConversationsList(ctx context.Context, opts ConversationsListOpts) ([]Channel, string, error) {
	err := checkPermission(ctx, "bot", "channels:read", "groups:read", "im:read", "mpim:read")
	if err != nil {
		return nil, "", err
	}
	inclArchived := !opts.ExcludeArchived
	inclPrivate := strings.Contains(opts.Types, "private_channel")
	inclPublic := strings.Contains(opts.Types, "public_channel") || opts.Types == ""

	if inclPublic && !hasScope(ctx, "bot", "channels:read") {
		return nil, "", &response{Err: "invalid_types"}
	}
	if inclPrivate && !hasScope(ctx, "bot", "groups:read") {
		return nil, "", &response{Err: "invalid_types"}
	}

	isBot := botID(ctx) != ""
	uid := userID(ctx)
	var cursorID string
	if opts.Cursor != "" {
		data, err := base64.URLEncoding.DecodeString(opts.Cursor)
		if err != nil {
			return nil, "", &response{Err: "invalid_cursor"}
		}
		cursorID = string(data)
		opts.Cursor = ""
	}
	filter := func(ch *channelState) bool {
		if ch == nil {
			return false
		}
		if cursorID != "" && cursorID >= ch.ID {
			return false
		}
		if ch.IsArchived && !inclArchived {
			return false
		}

		if ch.IsGroup && !inclPrivate {
			return false
		}
		if !ch.IsGroup && !inclPublic {
			return false
		}

		if ch.IsGroup && !isBot && !contains(ch.Users, uid) {
			return false
		}

		return true
	}

	if opts.Limit == 0 {
		opts.Limit = 100
	}
	if opts.Limit > 1000 {
		return nil, "", &response{Err: "invalid_limit"}
	}

	st.mx.Lock()
	defer st.mx.Unlock()

	ids := make([]string, 0, len(st.channels))
	for id := range st.channels {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	result := make([]Channel, 0, len(ids))
	for _, id := range ids {
		ch := st.channels[id]
		if !filter(ch) {
			continue
		}
		result = append(result, ch.Channel)
	}

	originalTotal := len(result)
	if len(result) > opts.Limit {
		result = result[:opts.Limit]
	}

	if len(result) > 1 {
		// limit is never guaranteed (only as max) as per the docs
		// so ensure it's handled by randomizing number of returned items
		max := rand.Intn(len(result)) + 1
		result = result[:max]
	}

	if originalTotal > len(result) && len(result) > 0 {
		opts.Cursor = base64.URLEncoding.EncodeToString([]byte(result[len(result)-1].ID))
	}

	return result, opts.Cursor, nil
}

// ServeConversationsList serves a request to the `conversations.list` API call.
//
// https://api.slack.com/methods/conversations.list
func (s *Server) ServeConversationsList(w http.ResponseWriter, req *http.Request) {
	var limit int
	limitStr := req.FormValue("limit")
	var err error
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			respondWith(w, &response{Err: "invalid_limit"})
			return
		}
	}

	chans, cur, err := s.API().ConversationsList(req.Context(), ConversationsListOpts{
		Cursor:          req.FormValue("cursor"),
		Limit:           limit,
		Types:           req.FormValue("types"),
		ExcludeArchived: req.FormValue("exclude_archived") == "true",
	})
	if respondErr(w, err) {
		return
	}

	var resp struct {
		response
		Channels []Channel `json:"channels"`
	}

	resp.Meta.Cursor = cur
	resp.Channels = chans
	resp.OK = true

	respondWith(w, resp)
}
