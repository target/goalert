package message

import "time"

// ThrottleRule sets the number of messages allowed to be sent per set duration.
type ThrottleRule struct {
	Count int
	Per   time.Duration

	// If Smooth is set, the rule will spread the count limit over the duration.
	Smooth bool
}

// ThrottleRules is a collection of ThrottleRule that implements the ThrottleConfig interface.
type ThrottleRules []ThrottleRule

// Rules always returns the set of configured rules for all messages.
func (rs ThrottleRules) Rules(Message) []ThrottleRule { return rs }

// MaxDuration returns the longest `Per` value for the set of rules.
func (rs ThrottleRules) MaxDuration() time.Duration {
	var dur time.Duration
	for _, r := range rs {
		if r.Per > dur {
			dur = r.Per
		}
	}
	return dur
}
