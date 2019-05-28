package processinglock

import (
	"fmt"

	"go.opencensus.io/trace"
)

// Config defines the parameters of the lock.
type Config struct {
	Type    Type
	Version int // Version must match the value in engine_processing_versions exactly or no lock will be obtained.
}

// String returns the string representation of Config.
func (cfg Config) String() string {
	return fmt.Sprintf("%s:v%d", cfg.Type, cfg.Version)
}

func (cfg Config) spanAttrs(extra ...trace.Attribute) []trace.Attribute {
	return append([]trace.Attribute{
		trace.StringAttribute("processingLock.type", string(cfg.Type)),
		trace.Int64Attribute("processingLock.version", int64(cfg.Version)),
	}, extra...)
}
func (cfg Config) decorateSpan(sp *trace.Span) {
	if sp == nil {
		return
	}

	sp.AddAttributes(cfg.spanAttrs()...)
}
