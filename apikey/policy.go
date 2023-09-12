package apikey

// GQLPolicy is a GraphQL API key policy.
type GQLPolicy struct {
	Version       int
	AllowedFields []string
}
