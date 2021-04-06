package slack

import (
	"context"
	"fmt"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
)

// 4. given a "code" field that expires after 10m
// 5. call oath.v2.access method with code
//   6. `curl -F code=1234 -F client_id=3336676.569200954261 -F client_secret=ABCDEFGH https://slack.com/api/oauth.v2.access`
// 7. token is returned under `authed_user.access_token`
// 8. store token in database with userID relation
// 9. redirect user to slack:// uri?
//
// notes:
// - oath tokens do not expire
// - provide a user_scope parameter with requested user scopes instead of, or in addition to, the scope parameter
func (h *Handler) ServeUserAuthCallback(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	cfg := config.FromContext(ctx)
	code := req.FormValue("code")
	uri := cfg.General.PublicURL + "/api/v2/slack/auth"
	resp, err := slack.GetOAuthV2ResponseContext(ctx, http.DefaultClient, cfg.Slack.ClientID, cfg.Slack.ClientSecret, code, uri)
	if err != nil {
		panic(err)
	}

	userID := permission.UserID(ctx)
	permission.SudoContext(req.Context(), func(ctx context.Context) {
		_, err := h.c.NotificationStore.InsertSlackUser(ctx, resp.Team.ID, resp.AuthedUser.ID, userID, resp.AuthedUser.AccessToken)
		if err != nil {
			panic(err)
		}
	})

	// attempt to delete original auth msg within slack
	meta, err := h.c.NotificationStore.FindUserAuthMessageData(ctx, resp.AuthedUser.ID) // todo: this function always returning NoRowsInResultSet
	if err != nil {
		fmt.Println("FindUserAuthMessageData failing")
		panic(err)
	}
	var api = slack.New(cfg.Slack.AccessToken)
	fmt.Println("attempting to delete ephemeral message")
	slackChan, ts, err := api.DeleteMessageContext(ctx, meta.ChannelID, meta.Timestamp)
	if err != nil {
		panic(err)
	}
	fmt.Println("channel resp: ", slackChan)
	fmt.Println("timestamp resp: ", ts)

	// todo: complete action
	// todo: redirect to slack:// channel somehow (or close browser tab)?
}
