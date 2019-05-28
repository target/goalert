package alert

import (
	"database/sql/driver"
	"fmt"
	"github.com/target/goalert/validation/validate"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// DedupType represents a type of dedup identifier.
type DedupType string

// DedupType can be auto or user-generated.
const (
	DedupTypeUser      = DedupType("user")
	DedupTypeAuto      = DedupType("auto")
	DedupTypeHeartbeat = DedupType("heartbeat")
)

// DedupID represents a de-duplication ID for alerts.
type DedupID struct {
	Type    DedupType
	Version int
	Payload string
}

// ParseDedupString will parse a string into a DedupID struct.
func ParseDedupString(s string) (*DedupID, error) {
	parts := strings.SplitN(s, ":", 3)
	if len(parts) != 3 {
		return nil, errors.New("invalid format")
	}
	vers, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}
	return &DedupID{
		Type:    DedupType(parts[0]),
		Version: vers,
		Payload: parts[2],
	}, nil
}

// Value implements the driver.Valuer interface.
func (d DedupID) Value() (driver.Value, error) {
	return fmt.Sprintf("%s:%d:%s", d.Type, d.Version, d.Payload), nil
}

// Scan implements the sql.Scanner interface.
func (d *DedupID) Scan(value interface{}) error {
	var parsed *DedupID
	var err error
	switch t := value.(type) {
	case []byte:
		parsed, err = ParseDedupString(string(t))
	case string:
		parsed, err = ParseDedupString(t)
	case nil:
		return errors.New("can't scan nil dedup id")
	default:
		return errors.Errorf("could not scan unknown type for DedupID(%T)", t)
	}
	if err != nil {
		return err
	}

	*d = *parsed
	return nil
}

// NewUserDedup will create a new DedupID from a user-provided string.
func NewUserDedup(str string) *DedupID {
	str = validate.SanitizeText(str, 512)
	if str == "" {
		return nil
	}
	return &DedupID{
		Type:    DedupTypeUser,
		Version: 1,
		Payload: str,
	}
}
