package apikey

import "github.com/target/goalert/permission"

// GQLPolicy is a GraphQL API key policy.
//
// Any changes to this MUST require the version to be incremented.
//
// If new fields are added, for example, they MUST be set to `omitempty`
// to ensure existing keys don't break.
//
// It MUST be possible to unmarshal & re-marshal the policy without changing the data for existing keys.
type GQLPolicy struct {
	Version int
	Query   string
	Role    permission.Role
}
