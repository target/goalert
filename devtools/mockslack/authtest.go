package mockslack

import (
	"net/http"
)

// ServeAuthTest serves a request to the `auth.test` API call.
//
// https://slack.com/api/auth.test
func (s *Server) ServeAuthTest(w http.ResponseWriter, req *http.Request) {
	var respData struct {
		OK                  bool   `json:"ok"`
		URL                 string `json:"url"`
		Team                string `json:"team"`
		User                string `json:"user"`
		TeamID              string `json:"team_id"`
		UserID              string `json:"user_id"`
		BotID               string `json:"bot_id"`
		IsEnterpriseInstall bool   `json:"is_enterprise_install"`
	}
	respData.OK = true
	respData.TeamID = s.teamID
	respondWith(w, respData)
}
