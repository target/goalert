package gadb

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// UIKConfig stores the configuration for an integration key for how to handle incoming requests.
type UIKConfig struct {
	Version int
	V1      UIKConfigV1
}

// Scan implements the Scanner interface.
func (cfg *UIKConfig) Scan(value interface{}) error {
	switch v := value.(type) {
	case json.RawMessage:
		err := json.Unmarshal(v, cfg)
		if err != nil {
			return err
		}
	case []byte:
		err := json.Unmarshal(v, cfg)
		if err != nil {
			return err
		}
	case string:
		err := json.Unmarshal([]byte(v), cfg)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported scan for DestV1 type: %T", value)
	}

	return nil
}

// Value implements the driver Valuer interface.
func (cfg UIKConfig) Value() (interface{}, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(data), nil
}

// UIKConfigV1 stores the configuration for an integration key for how to handle incoming requests.
type UIKConfigV1 struct {
	Rules []UIKRuleV1

	// DefaultActions are the actions to take if no rules match.
	DefaultActions []UIKActionV1
}

// UIKRuleV1 is a set of conditions and actions to take if those conditions are met.
type UIKRuleV1 struct {
	ID            uuid.UUID
	Name          string
	Description   string
	ConditionExpr string
	Actions       []UIKActionV1

	ContinueAfterMatch bool
}

// UIKActionV1 is a single action to take if a rule matches.
type UIKActionV1 struct {
	ChannelID uuid.UUID
	Dest      DestV1

	// Params are parameters that are determined at runtime (e.g., the message to send).
	// The keys are the parameter names, and the values are the Expr expression strings.
	Params map[string]string
}

func (act UIKActionV1) Param(name string) string {
	if act.Params == nil {
		return ""
	}
	return act.Params[name]
}

func (act *UIKActionV1) SetParam(name, value string) {
	if act.Params == nil {
		act.Params = make(map[string]string)
	}
	act.Params[name] = value
}
