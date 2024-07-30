package message

import (
	"time"

	"github.com/target/goalert/notification"
)

// ThrottleConfigBuilder can be used to build advanced throttle configurations.
type ThrottleConfigBuilder struct {
	parent *ThrottleConfigBuilder

	msgTypes []notification.MessageType
	dstTypes []string

	rules []builderRules
	max   time.Duration
}

func (b *ThrottleConfigBuilder) top() *ThrottleConfigBuilder {
	if b.parent != nil {
		return b.parent
	}
	return b
}

// WithMsgTypes allows adding rules for messages matching at least one MessageType.
func (b *ThrottleConfigBuilder) WithMsgTypes(msgTypes ...notification.MessageType) *ThrottleConfigBuilder {
	return &ThrottleConfigBuilder{
		parent: b.top(),

		msgTypes: msgTypes,
		dstTypes: b.dstTypes,
	}
}

// WithDestTypes allows adding rules for messages matching at least one DestType.
func (b *ThrottleConfigBuilder) WithDestTypes(destTypes ...string) *ThrottleConfigBuilder {
	return &ThrottleConfigBuilder{
		parent: b.top(),

		msgTypes: b.msgTypes,
		dstTypes: destTypes,
	}
}

func (b *ThrottleConfigBuilder) setMax(rules []ThrottleRule) {
	for _, r := range rules {
		if r.Per > b.max {
			b.max = r.Per
		}
	}
}

// AddRules will append a set of rules for the current filter (if any).
func (b *ThrottleConfigBuilder) AddRules(rules []ThrottleRule) {
	b.top().rules = append(b.top().rules, builderRules{
		msgTypes: b.msgTypes,
		dstTypes: b.dstTypes,
		rules:    rules,
	})
	b.top().setMax(rules)
}

// Config will return a ThrottleConfig for the current top-level configuration.
func (b *ThrottleConfigBuilder) Config() ThrottleConfig {
	var cfg builderConfig
	// copy rules
	cfg.rules = append(cfg.rules, b.top().rules...)
	cfg.max = b.max

	return &cfg
}

type builderRules struct {
	msgTypes []notification.MessageType
	dstTypes []string
	rules    []ThrottleRule
}

func (r builderRules) match(msg Message) bool {
	typeMatch := len(r.msgTypes) == 0
	for _, typ := range r.msgTypes {
		if typ != msg.Type {
			continue
		}

		typeMatch = true
		break
	}
	if !typeMatch {
		return false
	}

	destMatch := len(r.dstTypes) == 0
	for _, dst := range r.dstTypes {
		if dst != msg.Dest.ToDestV1().Type {
			continue
		}

		destMatch = true
		break
	}

	return destMatch
}

type builderConfig struct {
	rules []builderRules
	max   time.Duration
}

func (cfg *builderConfig) MaxDuration() time.Duration { return cfg.max }
func (cfg *builderConfig) Rules(msg Message) []ThrottleRule {
	var rules []ThrottleRule
	for _, set := range cfg.rules {
		if !set.match(msg) {
			continue
		}
		rules = append(rules, set.rules...)
	}

	return rules
}
