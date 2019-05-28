package app

import (
	"context"
	"net/http"

	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/twilio"

	"github.com/pkg/errors"
	"go.opencensus.io/plugin/ochttp"
)

func (app *App) initTwilio(ctx context.Context) error {
	app.twilioConfig = &twilio.Config{
		APIURL: app.cfg.TwilioBaseURL,
		Client: &http.Client{Transport: &ochttp.Transport{}},
	}

	var err error
	app.twilioSMS, err = twilio.NewSMS(ctx, app.db, app.twilioConfig)
	if err != nil {
		return errors.Wrap(err, "init TwilioSMS")
	}
	app.notificationManager.RegisterSender(notification.DestTypeSMS, "Twilio-SMS", app.twilioSMS)

	app.twilioVoice, err = twilio.NewVoice(ctx, app.db, app.twilioConfig)
	if err != nil {
		return errors.Wrap(err, "init TwilioVoice")
	}
	app.notificationManager.RegisterSender(notification.DestTypeVoice, "Twilio-Voice", app.twilioVoice)

	return nil
}
