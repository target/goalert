package app

import (
	"context"

	"github.com/target/goalert/expflag"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/slack"
)

func (app *App) initSlack(ctx context.Context) error {
	var err error
	app.slackChan, err = slack.NewChannelSender(ctx, slack.Config{
		BaseURL:   app.cfg.SlackBaseURL,
		UserStore: app.UserStore,
	})
	if err != nil {
		return err
	}
	app.notificationManager.RegisterSender(notification.DestTypeSlackChannel, "Slack-Channel", app.slackChan)
	app.notificationManager.RegisterSender(notification.DestTypeSlackDM, "Slack-DM", app.slackChan.DMSender())
	if expflag.ContextHas(ctx, expflag.SlackUserGroups) {
		app.notificationManager.RegisterSender(notification.DestTypeSlackUG, "Slack-UserGroup", app.slackChan.UserGroupSender())
	}

	return nil
}
