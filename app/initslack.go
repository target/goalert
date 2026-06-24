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
		Client:    app.httpClient,
	})
	if err != nil {
		return err
	}

	return nil
}
