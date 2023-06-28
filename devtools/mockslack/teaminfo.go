package mockslack

import (
	"context"
	"net/http"
)

// TeamInfoOpts contains parameters for the TeamInfo API call.
type TeamInfoOpts struct {
	Team string
}

type Team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (st *API) TeamInfo(ctx context.Context, id string) (*Team, error) {
	st.mx.Lock()
	defer st.mx.Unlock()
	if id != "" && id != st.teamID {
		return nil, &response{Err: "team_not_found"}
	}

	return &Team{
		ID:   st.teamID,
		Name: "Mock Slack Team",
	}, nil
}

// ServeTeamInfo serves a request to the `team.info` API call.
//
// https://api.slack.com/methods/team.info
func (s *Server) ServeTeamInfo(w http.ResponseWriter, req *http.Request) {
	t, err := s.API().TeamInfo(req.Context(), req.FormValue("team"))
	if respondErr(w, err) {
		return
	}

	var resp struct {
		response
		Team *Team `json:"team"`
	}
	resp.OK = true
	resp.Team = t

	respondWith(w, resp)
}
