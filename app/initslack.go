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
	app.notificationManager.RegisterSender(slack.DestTypeChannel, "Slack-Channel", app.slackChan)
	app.notificationManager.RegisterSender(slack.DestTypeDM, "Slack-DM", app.slackChan.DMSender())
	app.notificationManager.RegisterSender(slack.DestTypeUsergroup, "Slack-UserGroup", app.slackChan.UserGroupSender())

	return nil
}
