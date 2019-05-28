package mockslack

import (
	"context"
	"net/http"
)

// AuthRevoke will revoke the auth token from the provided context.
func (st *API) AuthRevoke(ctx context.Context, test bool) (bool, error) {
	st.mx.Lock()
	defer st.mx.Unlock()
	id := tokenID(ctx)
	tok := st.tokens[id]
	if tok == nil {
		return false, &response{Err: "invalid_auth"}
	}

	if !test {
		delete(st.tokens, id)
		return true, nil
	}

	return false, nil
}

// ServeAuthRevoke implements the auth.revoke API call.
//
// https://api.slack.com/methods/auth.revoke
func (s *Server) ServeAuthRevoke(w http.ResponseWriter, req *http.Request) {
	revoked, err := s.API().AuthRevoke(req.Context(), req.FormValue("test") != "")
	if respondErr(w, err) {
		return
	}

	var resp struct {
		response
		Revoked bool `json:"revoked"`
	}
	resp.OK = true
	resp.Revoked = revoked

	respondWith(w, resp)
}
