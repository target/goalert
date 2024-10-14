package processinglock

import (
	"fmt"
)

// Config defines the parameters of the lock.
type Config struct {
	Type    Type
	Version int32 // Version must match the value in engine_processing_versions exactly or no lock will be obtained.
}

// String returns the string representation of Config.
func (cfg Config) String() string {
	return fmt.Sprintf("%s:v%d", cfg.Type, cfg.Version)
}
