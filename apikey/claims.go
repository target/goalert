package apikey

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

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
