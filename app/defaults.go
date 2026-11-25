package app

import (
	"time"

	"github.com/target/goalert/notification/slack"
)

// Defaults returns the default app config.
func Defaults() Config {
	return Config{
		DBMaxOpen:         15,
		DBMaxIdle:         5,
		ListenAddr:        "localhost:8081",
		MaxReqBodyBytes:   256 * 1024,
		MaxReqHeaderBytes: 4096,
		RegionName:        "default",
		EngineCycleTime:   5 * time.Second,
		SMTPMaxRecipients: 1,

		SlackMaxChannelPages: slack.DefaultMaxChannelPages,
	}
}
