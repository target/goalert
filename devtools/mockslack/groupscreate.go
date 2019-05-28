package mockslack

import (
	"context"
	"net/http"
)

// GroupsCreate is used to create a channel.
func (st *API) GroupsCreate(ctx context.Context, opts ConversationCreateOpts) (*Channel, error) {
	err := checkPermission(ctx, "groups:write")
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
		ID:      st.gen.GroupID(),
		Name:    opts.Name,
		IsGroup: true,
	}

	st.mx.Lock()
	st.channels[ch.ID] = &channelState{Channel: ch}
	st.mx.Unlock()

	return &ch, nil
}

// ServeGroupsCreate serves a request to the `Groups.create` API call.
//
// https://api.slack.com/methods/Groups.create
func (s *Server) ServeGroupsCreate(w http.ResponseWriter, req *http.Request) {
	ch, err := s.API().GroupsCreate(req.Context(), ConversationCreateOpts{Name: req.FormValue("name"), Validate: req.FormValue("validate") == "true"})
	if respondErr(w, err) {
		return
	}

	var resp struct {
		response
		Group *Channel `json:"group"`
	}
	resp.OK = true
	resp.Group = ch

	respondWith(w, resp)
}
