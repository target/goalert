package slack

import (
	"context"
	"encoding/json"
	"net/http"
)

func lookupTeamIDForToken(ctx context.Context, token string) (string, error) {
	type RespBody struct {
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

	var body RespBody
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return "", err
	}
	return body.TeamID, nil
}
