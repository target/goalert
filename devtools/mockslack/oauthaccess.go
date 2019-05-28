package mockslack

import (
	"context"
	"net/http"
)

// OAuthAccessOpts contains parameters for an OAuthAccess API call.
type OAuthAccessOpts struct {
	ClientID     string
	ClientSecret string
	Code         string
}

// OAuthAccess will exchange a temporary code for an access token.
func (st *API) OAuthAccess(ctx context.Context, opts OAuthAccessOpts) (*AuthToken, error) {
	st.mx.Lock()
	defer st.mx.Unlock()

	app := st.apps[opts.ClientID]
	if app == nil {
		return nil, &response{Err: "invalid_client_id"}
	}

	if app.Secret != opts.ClientSecret {
		return nil, &response{Err: "bad_client_secret"}
	}

	tok := st.tokenCodes[opts.Code]
	if tok == nil || tok.ClientID != opts.ClientID {
		return nil, &response{Err: "invalid_code"}
	}

	delete(st.tokenCodes, opts.Code)

	return tok.AuthToken, nil
}

// ServeOAuthAccess serves a request to the `oauth.access` API call.
//
// https://api.slack.com/methods/oauth.access
func (s *Server) ServeOAuthAccess(w http.ResponseWriter, req *http.Request) {
	usr, pass, _ := req.BasicAuth()
	tok, err := s.API().OAuthAccess(req.Context(), OAuthAccessOpts{ClientID: usr, ClientSecret: pass, Code: req.FormValue("code")})
	if respondErr(w, err) {
		return
	}

	var resp struct {
		AccessToken string `json:"access_token"`
		UserID      string `json:"user_id"`
	}
	resp.AccessToken = tok.ID
	resp.UserID = tok.User

	respondWith(w, resp)
}
