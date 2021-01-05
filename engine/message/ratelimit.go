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

	// 1 per minute for all message types
	perCM.AddRules([]ThrottleRule{{Count: 1, Per: time.Minute}})

	// 1 per 15-minute for all status notifications
	perCM.
		WithMsgTypes(notification.MessageTypeAlertStatus, notification.MessageTypeAlertStatusBundle).
		AddRules([]ThrottleRule{{Count: 1, Per: 15 * time.Minute}})

	// alert-specific rules
	alerts := perCM.
		WithMsgTypes(notification.MessageTypeAlert, notification.MessageTypeAlertBundle)
	alerts.WithDestTypes(notification.DestTypeVoice).
		AddRules([]ThrottleRule{
			{Count: 1, Per: time.Minute},
			{Count: 3, Per: 15 * time.Minute},
			{Count: 7, Per: time.Hour, Smooth: true},
			{Count: 15, Per: 3 * time.Hour, Smooth: true},
		})
	alerts.WithDestTypes(notification.DestTypeSMS).
		AddRules([]ThrottleRule{
			{Count: 1, Per: time.Minute},
			{Count: 5, Per: 15 * time.Minute},
			{Count: 12, Per: time.Hour},
			{Count: 20, Per: 3 * time.Hour},
		})

	PerCMThrottle = perCM.Config()
}
