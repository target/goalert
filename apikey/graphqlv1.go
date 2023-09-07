package apikey

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type PolicyType string

const (
	PolicyTypeGraphQLV1 PolicyType = "graphql-v1"
)

type Policy struct {
	Type PolicyType

	GraphQLV1 *GraphQLV1 `json:",omitempty"`
}

type Type string

const (
	TypeGraphQLV1 Type = "graphql-v1"
)

type V1 struct {
	Type Type

	GraphQLV1 *GraphQLV1 `json:",omitempty"`
}

type GraphQLV1 struct {
	AllowedFields []GraphQLField `json:"f"`
}
type GraphQLField struct {
	ObjectName string `json:"o"`
	Name       string `json:"n"`
}

type Claims struct {
	jwt.RegisteredClaims
	PolicyHash []byte `json:"pol"`
}

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
