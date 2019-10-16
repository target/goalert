package message

import (
	"time"

	"github.com/target/goalert/app/lifecycle"
	"github.com/target/goalert/notification"
)

// RateConfig allows setting egress rate limiting on messages.
type RateConfig struct {
	// PerSecond determines the target messages-per-second limit.
	PerSecond int

	// Batch sets how often granularity of the rate limit.
	Batch time.Duration
}

// Config is used to configure the message sender.
type Config struct {
	// MaxMessagesPerCycle determines the number of pending messages
	// fetched per-cycle for delivery.
	//
	// Defaults to 50.
	MaxMessagesPerCycle int

	// RateLimit allows configuring rate limits per contact-method type.
	RateLimit map[notification.DestType]*RateConfig

	// Pausable is optional, and allows early-abort of
	// message sending when IsPaused returns true.
	Pausable lifecycle.Pausable
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		RateLimit: map[notification.DestType]*RateConfig{
			notification.DestTypeSMS: &RateConfig{
				PerSecond: 1,
				Batch:     5 * time.Second,
			},
			notification.DestTypeVoice: &RateConfig{
				PerSecond: 1,
				Batch:     5 * time.Second,
			},
		},
	}
}
