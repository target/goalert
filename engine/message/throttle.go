package message

import (
	"crypto/sha256"
	"time"

	"github.com/target/goalert/gadb"
)

// Throttle represents the throttled messages for a queue.
type Throttle struct {
	cfg      ThrottleConfig
	typeOnly bool
	now      time.Time

	first    map[ThrottleItem]time.Time
	count    map[ThrottleItem]int
	cooldown map[gadb.DestHash]bool
}

// ThrottleItem represents the messages being throttled.
type ThrottleItem struct {
	DestHash  gadb.DestHash
	BucketDur time.Duration
}

// ThrottleConfig provides ThrottleRules for a given message.
type ThrottleConfig interface {
	Rules(Message) []ThrottleRule
	MaxDuration() time.Duration
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

// NewThrottle creates a new Throttle used to manage outgoing messages in a queue.
func NewThrottle(cfg ThrottleConfig, now time.Time, byTypeOnly bool) *Throttle {
	return &Throttle{
		cfg:      cfg,
		now:      now,
		typeOnly: byTypeOnly,

		first:    make(map[ThrottleItem]time.Time),
		count:    make(map[ThrottleItem]int),
		cooldown: make(map[gadb.DestHash]bool),
	}
}

func (tr *Throttle) destKey(d gadb.DestV1) gadb.DestHash {
	if tr.typeOnly {
		return sha256.Sum256([]byte(d.Type))
	}

	return d.DestHash()
}

// Record keeps track of the outgoing messages being throttled in a queue.
func (tr *Throttle) Record(msg Message) {
	keyHash := tr.destKey(msg.Dest)

	since := tr.now.Sub(msg.SentAt)
	rules := tr.cfg.Rules(msg)
	for i, rule := range rules {
		if since >= rule.Per {
			continue
		}
		key := ThrottleItem{DestHash: keyHash, BucketDur: rule.Per}
		tr.count[key]++
		count := tr.count[key]
		if tr.first[key].IsZero() || msg.SentAt.Before(tr.first[key]) {
			tr.first[key] = msg.SentAt
		}

		if count >= rule.Count {
			tr.cooldown[keyHash] = true
			continue
		}

		if !rule.Smooth {
			continue
		}

		// flat rate
		var prevRule ThrottleRule
		if i > 0 {
			prevRule = rules[i-1]
		}

		if count < prevRule.Count || count == 0 {
			// allow prev rule in entirety
			continue
		}

		// spread remainder evenly
		count -= prevRule.Count
		elapsed := tr.now.Sub(tr.first[key]) - prevRule.Per
		per := rule.Per - prevRule.Per

		if count > int(elapsed*time.Duration(rule.Count-prevRule.Count)/per) {
			tr.cooldown[keyHash] = true
		}
	}
}

// InCooldown returns true or false depending on the cooldown state of a throttled message.
func (tr *Throttle) InCooldown(msg Message) bool {
	return tr.cooldown[tr.destKey(msg.Dest)]
}
