package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
)

// Handler responds to API requests from Slack
type Handler struct {
	c Config
}

// Payload represents the relevant payload information sent from Slack
type Payload struct {
	ResponseURL string `json:"response_url"`
	Actions     []Action
	Channel     slack.Channel
}

// Action represents the information given from an action event within Slack
// e.g. clicking to acknowledge an alert from slack
type Action struct {
	ActionID string `json:"action_id"`
	ActionTS string `json:"action_ts"`
	BlockID  string `json:"block_id"`
	Value    string
}

// NewHandler creates a new Handler, registering API routes using chi.
func NewHandler(c Config) *Handler {
	return &Handler{c: c}
}

func httpErr(w http.ResponseWriter, err error) error {
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	panic(err)
}

// validRequest is used to validate a request from Slack.
// If the request is validated true is returned, false otherwise.
// https://api.slack.com/authentication/verifying-requests-from-slack
func validRequest(w http.ResponseWriter, req *http.Request) error {
	if req.Method != "POST" {
		return httpErr(w, errors.New("not a post"))
	}

	secret := config.FromContext(req.Context()).Slack.SigningSecret
	sv, err := slack.NewSecretsVerifier(req.Header, secret)
	if err != nil {
		return err
	}

	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return httpErr(w, err)
	}
	req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	_, err = sv.Write(body)
	if err != nil {
		return httpErr(w, err)
	}

	return sv.Ensure()
}

func (h *Handler) processAction(ctx context.Context, w http.ResponseWriter, payload slack.InteractionCallback, action *slack.BlockAction, alertIDStr string) error {
	cfg := config.FromContext(ctx)
	var api = slack.New(cfg.Slack.AccessToken)

	// add source info to ctx to enable writing the action to alert log
	ncID, _, err := h.c.AlertLogStore.FindNCByValue(ctx, nil, payload.Channel.ID)
	if err != nil {
		return err
	}
	ctx = permission.UserSourceContext(ctx, payload.User.ID, permission.RoleUser, &permission.SourceInfo{
		Type: permission.SourceTypeNotificationChannel,
		ID:   ncID,
	})

	// handle button clicked within Slack
	alertID, err := strconv.Atoi(alertIDStr)
	if err != nil {
		return err
	}
	var actionErr error
	switch action.ActionID {
	case "ack":
		actionErr = h.c.AlertStore.UpdateStatus(ctx, alertID, alert.StatusActive)
	case "esc":
		actionErr = h.c.AlertStore.Escalate(ctx, alertID)
	case "close":
		actionErr = h.c.AlertStore.UpdateStatus(ctx, alertID, alert.StatusClosed)
	}
	if actionErr != nil {
		return err
	}

	a, err := h.c.AlertStore.FindOne(ctx, alertID)
	if err != nil {
		return err
	}
	msgOpt := CraftAlertMessage(*a, cfg.CallbackURL("/alerts/"+alertIDStr), payload.ResponseURL)

	// update original message in Slack
	_, _, err = api.PostMessageContext(ctx, payload.Channel.ID, msgOpt...)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	return nil
}

// ServeActionCallback processes POST requests from Slack. A callback ID is provided
// to determine which action to take.
func (h *Handler) ServeActionCallback(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Action received")
	err := validRequest(w, req)
	if err != nil {
		httpErr(w, err)
	}

	var payload slack.InteractionCallback
	err = json.Unmarshal([]byte(req.FormValue("payload")), &payload)
	if err != nil {
		httpErr(w, err)
	}

	process := func(ctx context.Context) {
		cfg := config.FromContext(ctx)
		var api = slack.New(cfg.Slack.AccessToken)

		// actions may come in batches, range over
		for _, action := range payload.ActionCallback.BlockActions {
			if action.ActionID == "openLink" {
				return
			}

			// check if user has linked Slack with their GoAlert account
			userID, err := h.c.AuthHandler.FindUserIDForAuthSubject(ctx, "slack:"+payload.Team.ID, payload.User.ID)
			if err != nil {
				fmt.Println("error finding user")
				httpErr(w, err)
			}

			// send Unauthorized message if user is not linked
			if userID == "" {
				_, err := api.PostEphemeralContext(ctx, payload.Channel.ID, payload.User.ID, needsAuthMsgOpt())
				if err != nil {
					fmt.Println("error posting ephemeral auth msg")
					httpErr(w, err)
				}
				return
			}

			alertIDStr := action.Value
			err = h.processAction(ctx, w, payload, action, alertIDStr)
			if err != nil {
				fmt.Println("error processing action")
				httpErr(w, err)
			}
		}
	}

	permission.SudoContext(req.Context(), process)
}
