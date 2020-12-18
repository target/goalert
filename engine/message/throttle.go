package message

import (
	"github.com/target/goalert/notification"
	"time"
)

// Throttle represents the throttled messages for a queue.
type Throttle struct {
	cfg      ThrottleConfig
	ignoreID bool
	now      time.Time

	count    map[ThrottleItem]int
	cooldown map[notification.Dest]bool
}

// ThrottleItem represents the messages being throttled.
type ThrottleItem struct {
	Dest      notification.Dest
	BucketDur time.Duration
}

// ThrottleConfig contains the ThrottleRules for each
// notification Destination Type (i.e. DestTypeVoice, DestTypeSMS, DestTypeSlackChannel).
type ThrottleConfig map[notification.DestType][]ThrottleRule

// MaxDuration returns the max duration set in the ThrottleConfig rules.
func (cfg ThrottleConfig) MaxDuration() time.Duration {
	var max time.Duration
	for _, rules := range cfg {
		for _, rule := range rules {
			if rule.Per > max {
				max = rule.Per
			}
		}
	}
	return max
}

func maxThrottleDuration(cfgs ...ThrottleConfig) time.Duration {
	var max time.Duration
	for _, cfg := range cfgs {
		dur := cfg.MaxDuration()
		if dur > max {
			max = dur
		}
	}
	return max
}

// ThrottleRules set the number of messages allowed to be sent per set duration.
type ThrottleRule struct {
	Count int
	Per   time.Duration
}

// NewThrottle creates a new Throttle used to manage outgoing messages in a queue.
func NewThrottle(cfg ThrottleConfig, now time.Time, ignoreID bool) *Throttle {
	return &Throttle{
		cfg:      cfg,
		now:      now,
		ignoreID: ignoreID,

		count:    make(map[ThrottleItem]int),
		cooldown: make(map[notification.Dest]bool),
	}
}

// Record keeps track of the outgoing messages being throttled in a queue.
func (tr *Throttle) Record(msg Message) {
	if tr.ignoreID {
		msg.Dest.ID = ""
	}
	msg.Dest.Value = ""

	since := tr.now.Sub(msg.SentAt)
	for _, rule := range tr.cfg[msg.Dest.Type] {
		if since >= rule.Per {
			continue
		}
		key := ThrottleItem{Dest: msg.Dest, BucketDur: rule.Per}
		tr.count[key]++
		count := tr.count[key]

		if count >= rule.Count {
			tr.cooldown[msg.Dest] = true
		}
	}
}

// InCooldown returns true or false depending on the cooldown state of a throttled message.
func (tr *Throttle) InCooldown(msg Message) bool {
	if tr.ignoreID {
		msg.Dest.ID = ""
	}
	msg.Dest.Value = ""

	return tr.cooldown[msg.Dest]
}
