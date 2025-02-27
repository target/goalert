package message

import (
	"time"

	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/email"
	"github.com/target/goalert/notification/twilio"
)

// GlobalCMThrottle represents the rate limits for each notification type.
var GlobalCMThrottle ThrottleConfig = ThrottleRules{{Count: 5, Per: 5 * time.Second}}

// PerCMThrottle configures rate limits for individual contact methods.
var PerCMThrottle ThrottleConfig

func init() {
	var perCM ThrottleConfigBuilder

	// Rate limit sms, voice and email types
	perCM.
		WithDestTypes(twilio.DestTypeTwilioVoice, twilio.DestTypeTwilioSMS, email.DestTypeEmail).
		AddRules([]ThrottleRule{{Count: 1, Per: time.Minute}})

	// On-Call Status Notifications
	perCM.
		WithMsgTypes(notification.MessageTypeScheduleOnCallUsers).
		AddRules([]ThrottleRule{
			{Count: 3, Per: 1 * time.Minute},
			{Count: 20, Per: 1 * time.Hour, Smooth: true},
		})

	// status notifications
	perCM.
		WithMsgTypes(notification.MessageTypeAlertStatus).
		WithDestTypes(twilio.DestTypeTwilioVoice, twilio.DestTypeTwilioSMS, email.DestTypeEmail).
		AddRules([]ThrottleRule{
			{Count: 1, Per: 3 * time.Minute},
			{Count: 3, Per: 20 * time.Minute},
			{Count: 8, Per: 120 * time.Minute, Smooth: true},
		})

	// alert notifications
	alertMessages := perCM.WithMsgTypes(notification.MessageTypeAlert, notification.MessageTypeAlertBundle)

	alertMessages.
		WithDestTypes(twilio.DestTypeTwilioVoice).
		AddRules([]ThrottleRule{
			{Count: 3, Per: 15 * time.Minute},
			{Count: 7, Per: time.Hour, Smooth: true},
			{Count: 15, Per: 3 * time.Hour, Smooth: true},
		})

	alertMessages.
		WithDestTypes(twilio.DestTypeTwilioSMS).
		AddRules([]ThrottleRule{
			{Count: 5, Per: 15 * time.Minute},
			{Count: 11, Per: time.Hour, Smooth: true},
			{Count: 21, Per: 3 * time.Hour, Smooth: true},
		})

	PerCMThrottle = perCM.Config()
}
