package mockslack

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type actionItem struct {
	ActionID string `json:"action_id"`
	BlockID  string `json:"block_id"`
	Text     struct {
		Type string
		Text string
	}
	Value string
	Type  string
}
type actionBody struct {
	Type    string
	AppID   string `json:"api_app_id"`
	Channel struct{ ID string }
	User    struct {
		ID       string
		Username string
		Name     string
		TeamID   string `json:"team_id"`
	}
	ResponseURL string `json:"response_url"`
	Actions     []actionItem
}

func (s *Server) ServeActionResponse(w http.ResponseWriter, r *http.Request) {
	actData := r.URL.Query().Get("action")
	var p actionBody
	err := json.Unmarshal([]byte(actData), &p)
	if respondErr(w, err) {
		return
	}
	r.Form.Set("channel", p.Channel.ID)
	s.ServeChatPostMessage(w, r)
}

// PerformActionAs will preform the action as the given user.
func (s *Server) PerformActionAs(userID string, a Action) error {
	usr := s.user(userID)
	if usr == nil {
		return errors.New("invalid Slack user ID")
	}

	app := s.app(a.AppID)
	if app == nil {
		return errors.New("invalid Slack app ID")
	}

	actionData, err := json.Marshal(a)
	if err != nil {
		return err
	}

	var p actionBody
	p.Type = "block_actions"
	p.User.ID = usr.ID
	p.User.Username = usr.Name
	p.User.Name = usr.Name
	p.User.TeamID = a.TeamID
	p.Channel.ID = a.ChannelID
	p.AppID = a.AppID
	p.ResponseURL = strings.TrimSuffix(s.urlPrefix, "/") + "/actions/response?action=" + url.QueryEscape(string(actionData))

	var action actionItem
	action.ActionID = a.ActionID
	action.BlockID = a.BlockID
	action.Text.Type = "plain_text"
	action.Text.Text = a.Text
	action.Value = a.Value
	action.Type = "button"
	p.Actions = append(p.Actions, action)

	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("serialize action payload: %w", err)
	}

	v := make(url.Values)
	v.Set("payload", string(data))

	resp, err := http.PostForm(app.ActionURL, v)
	if err != nil {
		return fmt.Errorf("perform action: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("perform action: %s", resp.Status)
	}

	return nil
}
