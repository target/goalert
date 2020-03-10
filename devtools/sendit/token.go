package sendit

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Token values
const (
	TokenIssuer          = "sendit"
	TokenAudienceAuth    = "auth"
	TokenAudienceConnect = "connect"
)

// GenerateToken will create a new token string with the given audience and subject,
// signed with the provided secret.
func GenerateToken(secret []byte, aud, sub string) (string, error) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		Issuer:    TokenIssuer,
		Audience:  aud,
		Subject:   sub,
		IssuedAt:  time.Now().Unix(),
		NotBefore: time.Now().Add(-15 * time.Minute).Unix(),
	})

	return tok.SignedString(secret)
}

// TokenSubject will return the subject of a token, after verifying the signature
// and audience.
func TokenSubject(secret []byte, aud, token string) (string, error) {
	tok, err := jwt.ParseWithClaims(token, &jwt.StandardClaims{}, func(tok *jwt.Token) (interface{}, error) {
		if tok.Method.Alg() != jwt.SigningMethodHS256.Name {
			return nil, jwt.ErrInvalidKeyType
		}
		return secret, nil
	})
	if err != nil {
		return "", err
	}

	claims := tok.Claims.(*jwt.StandardClaims)
	if !claims.VerifyIssuer(TokenIssuer, true) {
		return "", jwt.NewValidationError("invalid issuer", jwt.ValidationErrorIssuer)
	}
	if !claims.VerifyAudience(aud, true) {
		return "", jwt.NewValidationError("invalid audience", jwt.ValidationErrorAudience)
	}

	return claims.Subject, nil
}
