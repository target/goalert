package message

import (
	"github.com/target/goalert/notification"
	"time"
)

// PerCMThrottle represents the rate limits for each notification type per contact method.
var PerCMThrottle = ThrottleConfig{
	notification.DestTypeVoice: {
		{Count: 1, Per: time.Minute},
		{Count: 3, Per: 15 * time.Minute},
		{Count: 5, Per: time.Hour},
		{Count: 10, Per: 3 * time.Hour},
	},
	notification.DestTypeSMS: {
		{Count: 1, Per: time.Minute},
		{Count: 5, Per: 15 * time.Minute},
		{Count: 12, Per: time.Hour},
		{Count: 20, Per: 3 * time.Hour},
	},
	notification.DestTypeSlackChannel: {
		{Count: 1, Per: time.Minute},
	},
}

// GlobalCMThrottle represents the rate limits for each notification type.
var GlobalCMThrottle = ThrottleConfig{
	notification.DestTypeVoice: {
		{Count: 5, Per: 5 * time.Second},
	},
	notification.DestTypeSMS: {
		{Count: 5, Per: 5 * time.Second},
	},
	notification.DestTypeSlackChannel: {
		{Count: 5, Per: 5 * time.Second},
	},
}
