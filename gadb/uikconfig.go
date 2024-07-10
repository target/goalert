package gadb

import "github.com/google/uuid"

// UIKConfig stores the configuration for an integration key for how to handle incoming requests.
type UIKConfig struct {
	Version int
	V1      UIKConfigV1
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
