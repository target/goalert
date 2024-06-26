package apikey

import "github.com/target/goalert/permission"

// GQLPolicy is a GraphQL API key policy.
//
// Any changes to existing fields MUST require the version to be incremented.
//
// If new fields are added, they MUST be set to `omitempty`
// to ensure existing keys don't break.
//
// It MUST be possible to unmarshal & re-marshal the policy without changing the data for existing keys.
type GQLPolicy struct {
	Version int
	Query   string
	Role    permission.Role
}
