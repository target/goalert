package slack

import (
	"context"
	"encoding/json"
	"net/http"
)

func lookupTeamIDForToken(ctx context.Context, token string) (string, error) {
	type Meta struct {
		TeamID string `json:"team_id"`
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/auth.test", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var m Meta
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		return "", err
	}
	return m.TeamID, nil
}
