package app

import (
	"context"

	"github.com/target/goalert/notification/slack"
)

func (app *App) initSlack(ctx context.Context) error {
	var err error
	app.slackChan, err = slack.NewChannelSender(ctx, slack.Config{
		BaseURL:         app.cfg.SlackBaseURL,
		UserStore:       app.UserStore,
		Client:          app.httpClient,
		MaxChannelPages: app.cfg.SlackMaxChannelPages,
	})
	if err != nil {
		return err
	}

	return nil
}
