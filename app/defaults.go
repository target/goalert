package app

import "github.com/slack-go/slack"

// Defaults returns the default app config.
func Defaults() Config {
	return Config{
		DBMaxOpen:         15,
		DBMaxIdle:         5,
		ListenAddr:        "localhost:8081",
		MaxReqBodyBytes:   256 * 1024,
		MaxReqHeaderBytes: 4096,
		RegionName:        "default",
		TraceProbability:  0.01,

		SlackBaseURL: slack.APIURL,
	}
}
