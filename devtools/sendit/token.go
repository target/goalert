package sendit

import (
	"crypto/rand"
	"encoding/hex"
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

func genID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ValidPath will return true if `p` is a valid prefix path.
func ValidPath(p string) bool {
	if len(p) < 3 {
		return false
	}
	if len(p) > 64 {
		return false
	}
	for _, r := range p {
		if r == '-' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r >= 'a' && r <= 'z' {
			continue
		}
		return false
	}

	return true
}
