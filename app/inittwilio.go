package app

import (
	"context"

	"github.com/target/goalert/notification/twilio"

	"github.com/pkg/errors"
)

func (app *App) initTwilio(ctx context.Context) error {
	app.twilioConfig = &twilio.Config{
		BaseURL: app.cfg.TwilioBaseURL,
		CMStore: app.ContactMethodStore,
		DB:      app.db,
	}

	var err error
	app.twilioSMS, err = twilio.NewSMS(ctx, app.db, app.twilioConfig)
	if err != nil {
		return errors.Wrap(err, "init TwilioSMS")
	}
	app.notificationManager.RegisterSender(twilio.DestTypeTwilioSMS, "Twilio-SMS", app.twilioSMS)

	app.twilioVoice, err = twilio.NewVoice(ctx, app.db, app.twilioConfig)
	if err != nil {
		return errors.Wrap(err, "init TwilioVoice")
	}
	app.notificationManager.RegisterSender(twilio.DestTypeTwilioVoice, "Twilio-Voice", app.twilioVoice)

	return nil
}
