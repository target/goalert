package slack

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/slack-go/slack"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
)

// Handler responds to API requests from Slack
type Handler struct {
	c      Config
	respCh chan *notification.MessageResponse
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

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// validRequest is used to validate a request from Slack.
// If the request is validated true is returned, false otherwise.
// https://api.slack.com/authentication/verifying-requests-from-slack
func validRequest(w http.ResponseWriter, req *http.Request) bool {
	if req.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return false
	}

	ts := req.Header.Get("X-Slack-Request-Timestamp")

	// ignore request if more than 5 minutes from local time
	_ts, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return false
	}
	if abs(time.Now().Unix()-_ts) > 60*5 {
		return false
	}

	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return false
	}
	req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	secret := config.FromContext(req.Context()).Slack.SigningSecret
	h := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(h, "v0:%s:%s", ts, body)
	calculatedSignature := "v0=" + hex.EncodeToString(h.Sum(nil))
	signature := []byte(req.Header.Get("X-Slack-Signature"))

	return hmac.Equal(signature, []byte(calculatedSignature))
}

func (h *Handler) processAction(ctx context.Context, w http.ResponseWriter, action *slack.BlockAction, alertIDStr, channelID string) error {
	cfg := config.FromContext(ctx)
	var api = slack.New(cfg.Slack.AccessToken)

	// add source info to ctx to enable writing to alert log
	ncID, _, err := h.c.AlertLogStore.FindNCBySlackChanID(ctx, nil, channelID)
	if err != nil {
		return err
	}
	ctx = permission.SourceContext(ctx, &permission.SourceInfo{
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
	msgOpt := CraftAlertMessage(*a, cfg.CallbackURL("/alerts/"+alertIDStr))

	// if the alert was ever escalated, the alert may have multiple same-messages in for a given channel
	timestamps, err := h.c.NotificationStore.FindSlackAlertMsgTimestamps(ctx, alertID)
	if err != nil {
		return err
	}
	for _, ts := range timestamps {
		_, _, _, err := api.UpdateMessage(channelID, ts, msgOpt...)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
	}

	return nil
}

// ServeActionCallback processes POST requests from Slack. A callback ID is provided
// to determine which action to take.
func (h *Handler) ServeActionCallback(w http.ResponseWriter, req *http.Request) {
	if !validRequest(w, req) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var payload slack.InteractionCallback
	err := json.Unmarshal([]byte(req.FormValue("payload")), &payload)
	if err != nil {
		panic(err)
	}

	process := func(ctx context.Context) {
		cfg := config.FromContext(ctx)
		var api = slack.New(cfg.Slack.AccessToken)

		// actions may come in batches, range over
		for _, action := range payload.ActionCallback.BlockActions {
			alertIDStr := action.Value

			if action.ActionID == "auth" {
				_, err = h.c.NotificationStore.InsertUserAuthMetaData(ctx, payload.Team.ID, payload.User.ID, notification.UserAuthMetaData{
					ChannelID:   payload.Channel.ID,
					ResponseURL: payload.ResponseURL,
					AlertID:     alertIDStr,
				})
				if err != nil {
					http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
					return
				}
				return
			}
			if action.ActionID == "openLink" {
				return
			}

			// check if user valid, if ID does not exist return ephemeral to auth with GoAlert
			_, err = h.c.UserStore.FindOneBySlackUserID(ctx, payload.User.ID)
			if err != nil {
				uri := cfg.General.PublicURL + "/api/v2/slack/auth"
				msg := userAuthMessageOption(cfg.Slack.ClientID, alertIDStr, uri)
				_, err := api.PostEphemeralContext(ctx, payload.Channel.ID, payload.User.ID, msg)
				if err != nil {
					http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
					return
				}
				return
			}

			err = h.processAction(ctx, w, action, alertIDStr, payload.Channel.ID)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				panic(err)
			}
		}
	}

	permission.SudoContext(req.Context(), process)
}

// ListenResponse will return a channel that is fed async message responses.
func (h *Handler) ListenResponse() <-chan *notification.MessageResponse { return h.respCh }
