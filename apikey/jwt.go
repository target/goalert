package apikey

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Issuer is the JWT issuer for GraphQL API keys.
const Issuer = "goalert"

// Audience is the JWT audience for GraphQL API keys.
const Audience = "apikey-v1/graphql-v1"

// Claims is the set of claims that are encoded into a JWT for a GraphQL API key.
type Claims struct {
	jwt.RegisteredClaims
	PolicyHash []byte `json:"pol"`
}

// NewGraphQLClaims returns a new Claims object for a GraphQL API key with the embedded policy hash.
func NewGraphQLClaims(id uuid.UUID, policyHash []byte, expires time.Time) jwt.Claims {
	n := time.Now()
	return &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   id.String(),
			ExpiresAt: jwt.NewNumericDate(expires),
			IssuedAt:  jwt.NewNumericDate(n),
			NotBefore: jwt.NewNumericDate(n.Add(-time.Minute)),
			Issuer:    Issuer,
			Audience:  []string{Audience},
		},
		PolicyHash: policyHash,
	}
}
