package message

import (
	"time"

	"github.com/target/goalert/notification"
)

/*
- Make these public and add comments
- Update queue test to pass (happy path, using throttle)
- Add unit tests for throttle methods

- Optimize message DB access
- Remove existing rate-limit config

- Look into final rule set (e.g. look at real-world use/data)
*/

type throttle struct {
	cfg      throttleConfig
	ignoreID bool
	now      time.Time

	count    map[throttleItem]int
	cooldown map[notification.Dest]bool
}

type throttleItem struct {
	Dest      notification.Dest
	BucketDur time.Duration
}

type throttleConfig map[notification.DestType][]throttleRule

type throttleRule struct {
	Count int
	Per   time.Duration
}

func newThrottle(cfg throttleConfig, now time.Time, ignoreID bool) *throttle {
	return &throttle{
		cfg:      cfg,
		now:      now,
		ignoreID: ignoreID,

		count:    make(map[throttleItem]int),
		cooldown: make(map[notification.Dest]bool),
	}
}

func (tr *throttle) Record(msg Message) {
	if tr.ignoreID {
		msg.Dest.ID = ""
	}
	msg.Dest.Value = ""

	since := tr.now.Sub(msg.SentAt)
	for _, rule := range tr.cfg[msg.Dest.Type] {
		if since >= rule.Per {
			continue
		}

		key := throttleItem{Dest: msg.Dest, BucketDur: rule.Per}
		tr.count[key]++
		count := tr.count[key]

		if count >= rule.Count {
			tr.cooldown[msg.Dest] = true
		}
	}
}
func (tr *throttle) InCooldown(msg Message) bool {
	if tr.ignoreID {
		msg.Dest.ID = ""
	}
	msg.Dest.Value = ""

	return tr.cooldown[msg.Dest]
}
