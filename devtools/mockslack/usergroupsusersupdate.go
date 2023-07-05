package mockslack

import (
	"context"
	"net/http"
	"strings"
)

type UserGroupsUsersUpdateOptions struct {
	Usergroup string
	Users     []string
}

// UserGroupsUsersUpdate updates the list of users within a user group and returns the user group.
func (st *API) UserGroupsUsersUpdate(ctx context.Context, opts UserGroupsUsersUpdateOptions) (*UserGroup, error) {
	err := checkPermission(ctx, "bot", "usergroups:write")
	if err != nil {
		return nil, err
	}

	st.mx.Lock()
	defer st.mx.Unlock()

	ug := st.usergroups[opts.Usergroup]
	if ug == nil {
		return nil, &response{Err: "subteam_not_found"}
	}

	if len(opts.Users) == 0 {
		return nil, &response{Err: "no_users_provided"}
	}

	for _, u := range opts.Users {
		if st.users[u] == nil {
			return nil, &response{Err: "invalid_users"}
		}
	}

	ug.Users = opts.Users

	return &ug.UserGroup, nil
}

// ServeUserGroupsUsersUpdate serves a request to the `usergroups.users.update` API call.
//
// https://api.slack.com/methods/usergroups.users.update
func (s *Server) ServeUserGroupsUsersUpdate(w http.ResponseWriter, req *http.Request) {
	ug, err := s.API().UserGroupsUsersUpdate(req.Context(), UserGroupsUsersUpdateOptions{Usergroup: req.FormValue("usergroup"), Users: strings.Split(req.FormValue("users"), ",")})
	if respondErr(w, err) {
		return
	}

	var resp struct {
		response
		UserGroup *UserGroup `json:"usergroup"`
	}

	resp.UserGroup = ug

	respondWith(w, resp)
}
