package message

import (
	"time"

	"github.com/target/goalert/notification"
)

// GlobalCMThrottle represents the rate limits for each notification type.
var GlobalCMThrottle ThrottleConfig = ThrottleRules{{Count: 5, Per: 5 * time.Second}}

// PerCMThrottle configures rate limits for individual contact methods.
var PerCMThrottle ThrottleConfig

func init() {
	var perCM ThrottleConfigBuilder

	// All message types minus Slack
	perCM.
		WithDestTypes(notification.DestTypeUnknown, notification.DestTypeVoice, notification.DestTypeSMS, notification.DestTypeUserEmail).
		AddRules([]ThrottleRule{{Count: 1, Per: time.Minute}})

	// Slack messages
	// https://api.slack.com/docs/rate-limits#rate-limits__limits-when-posting-messages
	perCM.WithDestTypes(notification.DestTypeSlackChannel).
		AddRules([]ThrottleRule{
			{Count: 1, Per: time.Second},
		})

	// Status update notifications
	perCM.
		WithMsgTypes(notification.MessageTypeAlertStatus, notification.MessageTypeAlertStatusBundle).
		WithDestTypes(notification.DestTypeVoice, notification.DestTypeSMS).
		AddRules([]ThrottleRule{
			{Count: 1, Per: 3 * time.Minute},
			{Count: 3, Per: 20 * time.Minute},
			{Count: 8, Per: 120 * time.Minute, Smooth: true},
		})

	// Alert notifications
	alertMessages := perCM.WithMsgTypes(notification.MessageTypeAlert, notification.MessageTypeAlertBundle)

	alertMessages.
		WithDestTypes(notification.DestTypeVoice).
		AddRules([]ThrottleRule{
			{Count: 3, Per: 15 * time.Minute},
			{Count: 7, Per: time.Hour, Smooth: true},
			{Count: 15, Per: 3 * time.Hour, Smooth: true},
		})

	alertMessages.
		WithDestTypes(notification.DestTypeSMS).
		AddRules([]ThrottleRule{
			{Count: 5, Per: 15 * time.Minute},
			{Count: 11, Per: time.Hour, Smooth: true},
			{Count: 21, Per: 3 * time.Hour, Smooth: true},
		})

	PerCMThrottle = perCM.Config()
}
