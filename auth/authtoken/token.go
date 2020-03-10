package authtoken

import (
	"encoding/base64"
	"time"

	uuid "github.com/satori/go.uuid"
)

var b64Encoding = base64.URLEncoding.WithPadding(base64.NoPadding)

// Token represents an authentication token.
type Token struct {
	Version   int
	Type      Type
	ID        uuid.UUID
	CreatedAt time.Time
}
