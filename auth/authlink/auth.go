package authlink

import (
	"context"

	"github.com/dgrijalva/jwt-go"
	"github.com/target/goalert/auth"
)

// Auth will validate and claim an auth token, returning the UserID if successful.
func (s *Store) Auth(ctx context.Context, token string) (string, error) {
	var jwtClaims jwt.StandardClaims
	_, err := s.keys.VerifyJWT(token, &jwtClaims)
	if err != nil {
		return "", err
	}
	if !jwtClaims.VerifyIssuer(issuer, true) {
		return "", auth.Error("invalid token")
	}
	if !jwtClaims.VerifyAudience(audience, true) {
		return "", auth.Error("invalid token")
	}

	var userID string
	err = s.auth.QueryRowContext(ctx, jwtClaims.Subject).Scan(&userID)
	return userID, err
}
