package mockslack

import (
	"context"
	"net/http"
)

// UsersInfoOpts contains parameters for the UsersInfo API call.
type UsersInfoOpts struct {
	User string
}

func (st *API) UsersInfo(ctx context.Context, id string) (*User, error) {
	err := checkPermission(ctx, "bot", "users:read")
	if err != nil {
		return nil, err
	}

	st.mx.Lock()
	defer st.mx.Unlock()

	u := st.users[id]
	if u == nil {
		return nil, &response{Err: "user_not_found"}
	}

	return &u.User, nil
}

// ServeUsersInfo serves a request to the `users.info` API call.
//
// https://api.slack.com/methods/users.info
func (s *Server) ServeUsersInfo(w http.ResponseWriter, req *http.Request) {
	u, err := s.API().UsersInfo(req.Context(), req.FormValue("user"))
	if respondErr(w, err) {
		return
	}

	var resp struct {
		response
		User *User `json:"user"`
	}
	resp.OK = true
	resp.User = u

	respondWith(w, resp)
}
