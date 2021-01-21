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

	// all message types
	perCM.AddRules([]ThrottleRule{{Count: 1, Per: time.Minute}})

	// status notifications
	perCM.
		WithMsgTypes(notification.MessageTypeAlertStatus, notification.MessageTypeAlertStatusBundle).
		AddRules([]ThrottleRule{
			{Count: 1, Per: 3 * time.Minute},
			{Count: 3, Per: 20 * time.Minute},
			{Count: 8, Per: 120 * time.Minute, Smooth: true},
		})

	// alert notifications
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
