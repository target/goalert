package apikey

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Type string

const (
	TypeGraphQLV1 Type = "graphql-v1"
)

type V1 struct {
	Type Type

	GraphQLV1 *GraphQLV1 `json:",omitempty"`
}

type GraphQLV1 struct {
	Query  string
	SHA256 [32]byte
}

type GraphQLClaims struct {
	jwt.RegisteredClaims
	AuthHash [32]byte `json:"q"`
}

func NewGraphQLClaims(id uuid.UUID, queryHash [32]byte, expires time.Time) jwt.Claims {
	n := time.Now()
	return &GraphQLClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   id.String(),
			ExpiresAt: jwt.NewNumericDate(expires),
			IssuedAt:  jwt.NewNumericDate(n),
			NotBefore: jwt.NewNumericDate(n.Add(-time.Minute)),
			Issuer:    "goalert",
			Audience:  []string{"apikey-v1/graphql-v1"},
		},
		AuthHash: queryHash,
	}
}
