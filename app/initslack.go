package app

import (
	"context"

	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/slack"
)

func (app *App) initSlack(ctx context.Context) error {
	var err error
	app.slackChan, err = slack.NewChannelSender(ctx, slack.Config{
		BaseURL: app.cfg.SlackBaseURL,
	})
	if err != nil {
		return err
	}
	app.notificationManager.RegisterSender(notification.DestTypeSlackChannel, "Slack-Channel", app.slackChan)
	return nil
}
