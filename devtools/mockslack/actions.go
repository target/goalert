package mockslack

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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
	Team struct {
		ID     string
		Domain string
	}
	ResponseURL string `json:"response_url"`
	Actions     []actionItem
}

func (s *Server) ServeActionResponse(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string
		Type string `json:"response_type"`

		Blocks []struct {
			Type     string
			Text     struct{ Text string }
			Elements []struct {
				Type     string
				Text     struct{ Text string }
				Value    string
				ActionID string `json:"action_id"`
				URL      string
			}
		}
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Type != "ephemeral" {
		http.Error(w, "unexpected response type", http.StatusBadRequest)
		return
	}

	actData := r.URL.Query().Get("action")
	var a Action
	err := json.Unmarshal([]byte(actData), &a)
	if respondErr(w, err) {
		return
	}

	opts := ChatPostMessageOptions{
		ChannelID: a.ChannelID,
		User:      r.URL.Query().Get("user"),
	}

	if len(req.Blocks) > 0 {
		// new API
		for _, block := range req.Blocks {
			switch block.Type {
			case "section":
				opts.Text = block.Text.Text
			case "actions":
				for _, action := range block.Elements {
					if action.Type != "button" {
						continue
					}

					opts.Actions = append(opts.Actions, Action{
						ChannelID: a.ChannelID,
						TeamID:    a.TeamID,
						AppID:     a.AppID,
						ActionID:  action.ActionID,
						Text:      action.Text.Text,
						Value:     action.Value,
						URL:       action.URL,
					})
				}
			}
		}
	} else {
		opts.Text = req.Text
	}

	msg, err := s.API().ChatPostMessage(r.Context(), opts)
	if respondErr(w, err) {
		return
	}

	var respData struct {
		response
		TS      string
		Channel string   `json:"channel"`
		Message *Message `json:"message"`
	}
	respData.TS = msg.TS
	respData.OK = true
	respData.Channel = msg.ChannelID
	respData.Message = msg

	respondWith(w, respData)
}

// PerformActionAs will perform the action as the given user.
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
	p.Team.ID = a.TeamID
	p.Team.Domain = "example.com"
	p.Channel.ID = a.ChannelID
	p.AppID = a.AppID

	tok := s.newToken(AuthToken{
		User: userID,

		Scopes: []string{"bot"},
	})
	p.ResponseURL = fmt.Sprintf("%s/actions/response?token=%s&user=%s&action=%s", strings.TrimSuffix(s.urlPrefix, "/"), url.QueryEscape(tok.ID), url.QueryEscape(usr.ID), url.QueryEscape(string(actionData)))

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
	data = []byte(v.Encode())

	req, err := http.NewRequest("POST", app.ActionURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create action request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	t := time.Now()
	req.Header.Set("X-Slack-Request-Timestamp", strconv.FormatInt(t.Unix(), 10))
	h := hmac.New(sha256.New, []byte(app.SigningSecret))
	fmt.Fprintf(h, "v0:%d:%s", t.Unix(), string(data))
	req.Header.Set("X-Slack-Signature", "v0="+fmt.Sprintf("%x", h.Sum(nil)))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform action: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("perform action: %s", resp.Status)
	}

	return nil
}
