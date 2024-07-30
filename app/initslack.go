package app

import (
	"context"

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
	app.notificationManager.RegisterSender(slack.DestTypeSlackChannel, "Slack-Channel", app.slackChan)
	app.notificationManager.RegisterSender(slack.DestTypeSlackDirectMessage, "Slack-DM", app.slackChan.DMSender())
	app.notificationManager.RegisterSender(slack.DestTypeSlackUsergroup, "Slack-UserGroup", app.slackChan.UserGroupSender())

	return nil
}
