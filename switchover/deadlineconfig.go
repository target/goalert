package switchover

import (
	"context"
	"net/url"
	"time"
)

type ctxValue int

const (
	ctxValueDeadlines ctxValue = iota
)

// DeadlineConfig controls the timeing of a Switch-Over operation.
type DeadlineConfig struct {
	BeginAt          time.Time     // The start-time of the Switch-Over.
	ConsensusTimeout time.Duration // Amount of time to wait for consensus amongst all nodes before aborting.
	PauseDelay       time.Duration // How long to wait after starting before beginning the global pause.
	PauseTimeout     time.Duration // Timeout to achieve global pause before aborting.
	MaxPause         time.Duration // Absolute maximum amount of time for any operation to be delayed due to the Switch-Over.
	NoPauseAPI       bool          // Allow HTTP/API requests during Pause phase.
}

// DefaultConfig returns the default deadline configuration.
func DefaultConfig() DeadlineConfig {
	return DeadlineConfig{
		ConsensusTimeout: 3 * time.Second,
		PauseDelay:       5 * time.Second,
		PauseTimeout:     10 * time.Second,
		MaxPause:         13 * time.Second,
	}
}

// ConfigFromContext returns the DeadlineConfig associated with the current context.
func ConfigFromContext(ctx context.Context) DeadlineConfig {
	d, _ := ctx.Value(ctxValueDeadlines).(DeadlineConfig)
	return d
}

// PauseDeadline will return the deadline to achieve global pause.
func (cfg DeadlineConfig) PauseDeadline() time.Time {
	return cfg.BeginAt.Add(cfg.PauseDelay + cfg.PauseTimeout)
}

// ConsensusDeadline will return the deadline for consensus amonst all nodes.
func (cfg DeadlineConfig) ConsensusDeadline() time.Time {
	return cfg.BeginAt.Add(cfg.ConsensusTimeout)
}

// PauseAt will return the time global pause begins.
func (cfg DeadlineConfig) PauseAt() time.Time {
	return cfg.BeginAt.Add(cfg.PauseDelay)
}

// AbsoluteDeadline will calculate the absolute deadline of the entire switchover operation.
func (cfg DeadlineConfig) AbsoluteDeadline() time.Time {
	return cfg.BeginAt.Add(cfg.PauseDelay + cfg.MaxPause)
}

// Serialize returns a textual representation of DeadlineConfig that can be
// transmitted to other nodes. Offset should be time difference between
// the local clock and the central clock (i.e. Postgres).
func (cfg DeadlineConfig) Serialize(offset time.Duration) string {
	v := make(url.Values)
	v.Set("BeginAt", cfg.BeginAt.Add(-offset).Format(time.RFC3339Nano))
	v.Set("ConsensusTimeout", cfg.ConsensusTimeout.String())
	v.Set("PauseDelay", cfg.PauseDelay.String())
	v.Set("PauseTimeout", cfg.PauseTimeout.String())
	v.Set("MaxPause", cfg.MaxPause.String())
	noPauseAPI := "false"
	if cfg.NoPauseAPI {
		noPauseAPI = "true"
	}
	v.Set("NoPauseAPI", noPauseAPI)
	return v.Encode()
}

// ParseDeadlineConfig will parse deadline configuration (given by Serialize) from a string.
// Offset should be the time difference between the local clock and the central clock (i.e. Postgres).
func ParseDeadlineConfig(s string, offset time.Duration) (*DeadlineConfig, error) {
	v, err := url.ParseQuery(s)
	if err != nil {
		return nil, err
	}
	begin, err := time.Parse(time.RFC3339Nano, v.Get("BeginAt"))
	if err != nil {
		return nil, err
	}
	p := func(name string) (dur time.Duration) {
		if err != nil {
			return dur
		}
		dur, err = time.ParseDuration(v.Get(name))
		return dur
	}

	return &DeadlineConfig{
		BeginAt:          begin.Add(offset),
		ConsensusTimeout: p("ConsensusTimeout"),
		PauseDelay:       p("PauseDelay"),
		PauseTimeout:     p("PauseTimeout"),
		MaxPause:         p("MaxPause"),
		NoPauseAPI:       v.Get("NoPauseAPI") == "true",
	}, err
}
