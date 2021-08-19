package slack

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
)

// Handler responds to API requests from Slack
type Handler struct {
	c Config
}

// NewHandler creates a new Handler, registering API routes using chi.
func NewHandler(c Config) *Handler {
	return &Handler{c: c}
}

// validRequest is used to validate a request from Slack.
// If the request is validated true is returned, false otherwise.
// https://api.slack.com/authentication/verifying-requests-from-slack
func validRequest(w http.ResponseWriter, req *http.Request) error {
	if req.Method != "POST" {
		return errors.New("not a post")
	}

	secret := config.FromContext(req.Context()).Slack.SigningSecret
	sv, err := slack.NewSecretsVerifier(req.Header, secret)
	if err != nil {
		return err
	}

	defer req.Body.Close()
	body, err := ioutil.ReadAll(io.TeeReader(req.Body, &sv))
	if err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return sv.Ensure()
}

// ServeActionCallback processes POST requests from Slack. A callback ID is provided
// to determine which action to take.
func (h *Handler) ServeActionCallback(w http.ResponseWriter, req *http.Request) {
	err := validRequest(w, req)
	if err != nil {
		errutil.HTTPError(req.Context(), w, err)
	}

	var payload slack.InteractionCallback
	err = json.Unmarshal([]byte(req.FormValue("payload")), &payload)
	if err != nil {
		errutil.HTTPError(req.Context(), w, err)
	}

	ctx := permission.UserSourceContext(req.Context(), payload.User.ID, permission.RoleUser, &permission.SourceInfo{
		Type: permission.SourceTypeNotificationCallback,
		ID:   "slack:" + payload.Team.ID + ":" + payload.User.ID,
	})
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
			errutil.HTTPError(ctx, w, err)
		}

		// send Unauthorized message if user is not linked
		if userID == "" {
			_, err := api.PostEphemeralContext(ctx, payload.Channel.ID, payload.User.ID, needsAuthMsgOpt())
			if err != nil {
				errutil.HTTPError(ctx, w, err)
			}
			return
		}

		alertIDStr := action.Value

		// add source info to ctx to enable writing the action to alert log
		ncID, _, err := h.c.AlertLogStore.FindNCByValue(ctx, nil, payload.Channel.ID)
		if err != nil {
			errutil.HTTPError(ctx, w, err)
		}
		ctx = permission.UserSourceContext(ctx, payload.User.ID, permission.RoleUser, &permission.SourceInfo{
			Type: permission.SourceTypeNotificationChannel,
			ID:   ncID,
		})

		// handle button clicked within Slack
		alertID, err := strconv.Atoi(alertIDStr)
		if err != nil {
			errutil.HTTPError(ctx, w, err)
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
			errutil.HTTPError(ctx, w, err)
		}

		a, err := h.c.AlertStore.FindOne(ctx, alertID)
		if err != nil {
			errutil.HTTPError(ctx, w, err)
		}
		msgOpt := CraftAlertMessage(*a, cfg.CallbackURL("/alerts/"+alertIDStr), payload.ResponseURL)

		// update original message in Slack
		_, _, err = api.PostMessageContext(ctx, payload.Channel.ID, msgOpt...)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
	}
}
