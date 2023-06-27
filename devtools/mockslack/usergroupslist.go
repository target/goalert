package mockslack

import (
	"context"
	"net/http"
	"sort"
)

// UserGroupList returns a list of User Groups in a workspace.
func (st *API) UserGroupList(ctx context.Context) ([]UserGroup, error) {
	err := checkPermission(ctx, "bot", "usergroups:read")
	if err != nil {
		return nil, err
	}

	st.mx.Lock()
	defer st.mx.Unlock()

	ids := make([]string, 0, len(st.usergroups))
	for id := range st.usergroups {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	result := make([]UserGroup, 0, len(ids))
	for _, id := range ids {
		ch := st.usergroups[id]
		result = append(result, ch.UserGroup)
	}

	return result, nil
}

// ServeUserGroupList serves a request to the `usergroups.list` API call.
//
// https://api.slack.com/methods/usergroups.list
func (s *Server) ServeUserGroupList(w http.ResponseWriter, req *http.Request) {
	ug, err := s.API().UserGroupList(req.Context())
	if respondErr(w, err) {
		return
	}

	var resp struct {
		response
		UserGroups []UserGroup `json:"usergroups"`
	}

	resp.UserGroups = ug

	respondWith(w, resp)
}
