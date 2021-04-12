package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/slack-go/slack"
	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
)

func (h *Handler) ServeUserAuthCallback(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	cfg := config.FromContext(ctx)
	code := req.FormValue("code")
	uri := cfg.General.PublicURL + "/api/v2/slack/auth"
	resp, err := slack.GetOAuthV2ResponseContext(ctx, http.DefaultClient, cfg.Slack.ClientID, cfg.Slack.ClientSecret, code, uri)
	if err != nil {
		panic(err)
	}

	// store user/slack relation
	userID := permission.UserID(ctx)
	permission.SudoContext(req.Context(), func(ctx context.Context) {
		_, err := h.c.NotificationStore.InsertSlackUser(ctx, resp.Team.ID, resp.AuthedUser.ID, userID, resp.AuthedUser.AccessToken)
		if err != nil {
			panic(err)
		}
	})

	// remove ephemeral "link to goalert" msg in slack
	// note/todo: limit is message ts < 30min ago
	meta, err := h.c.NotificationStore.FindUserAuthMetaData(ctx, resp.AuthedUser.ID)
	if err != nil {
		panic(err)
	}
	values := map[string]string{"response_type": "ephemeral", "text": "", "replace_original": "true", "delete_original": "true"}
	jsonValue, _ := json.Marshal(values)
	_, err = http.Post(meta.ResponseURL, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		panic(err)
	}

	// todo: complete original action

	// redirect to most recent alert msg in channel that the user authed against
	var api = slack.New(cfg.Slack.AccessToken)
	alertID, err := strconv.Atoi(meta.AlertID)
	if err != nil {
		panic(err)
	}
	timestamps, err := h.c.NotificationStore.FindSlackAlertMsgTimestamps(ctx, alertID)
	if err != nil {
		panic(err)
	}

	url, err := api.GetPermalinkContext(ctx, &slack.PermalinkParameters{
		Channel: meta.ChannelID,
		Ts:      timestamps[len(timestamps)-1],
	})
	if err != nil {
		panic(err)
	}
	http.Redirect(w, req, url, http.StatusPermanentRedirect)
}
