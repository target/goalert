package message

import (
	"github.com/target/goalert/app/lifecycle"
	"github.com/target/goalert/notification"
	"time"
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

// batchNum returns the maximum number of messages to be sent per-cycle for the given type.
// If there is no limit, 0 is returned.
func (c Config) batchNum(t notification.DestType) int {
	bDur := c.batch(t)
	if bDur == 0 {
		return 0
	}
	pSec := c.perSecond(t)
	if pSec == 0 {
		return 0
	}

	max := int(bDur.Seconds() * float64(pSec))
	return max
}

// perSecond returns the number of messages to send per-second.
func (c Config) perSecond(t notification.DestType) int {
	cfg := c.RateLimit[t]
	if cfg == nil {
		return 0
	}
	return cfg.PerSecond
}

// batch returns the duration for a batch of messages.
func (c Config) batch(t notification.DestType) time.Duration {
	cfg := c.RateLimit[t]
	if cfg == nil {
		return time.Duration(0)
	}
	return cfg.Batch
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
