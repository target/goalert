package mockslack

import (
	"net/http"
)

// ServeAuthTest serves a request to the `auth.test` API call.
//
// https://slack.com/api/auth.test
func (s *Server) ServeAuthTest(w http.ResponseWriter, req *http.Request) {
	var respData struct {
		TeamID string `json:"team_id"`
	}
	respData.TeamID = s.state.teamID
	respondWith(w, respData)
}
