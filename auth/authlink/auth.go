package authlink

import (
	"context"
	"database/sql"

	"github.com/dgrijalva/jwt-go"
	"github.com/target/goalert/auth"
)

// Auth will validate and claim an auth token, returning the UserID if successful.
func (s *Store) Auth(ctx context.Context, token string) (string, error) {
	var jwtClaims jwt.StandardClaims
	_, err := s.keys.VerifyJWT(token, &jwtClaims)
	if err != nil {
		return "", auth.Error("invalid token")
	}
	if !jwtClaims.VerifyIssuer(issuer, true) {
		return "", auth.Error("invalid token")
	}
	if !jwtClaims.VerifyAudience(audience, true) {
		return "", auth.Error("invalid token")
	}

	var userID string
	err = s.auth.QueryRowContext(ctx, jwtClaims.Subject).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", auth.Error("invalid token")
	}
	return userID, err
}
