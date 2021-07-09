package authtoken

import (
	"encoding/base64"
	"time"

	"github.com/google/uuid"
)

var b64Encoding = base64.URLEncoding.WithPadding(base64.NoPadding)

// Token represents an authentication token.
type Token struct {
	Version   int
	Type      Type
	ID        uuid.UUID
	CreatedAt time.Time
}
