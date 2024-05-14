package authtoken

import (
	"encoding/base64"
	"encoding/binary"

	"github.com/target/goalert/validation"
)

// A SignFunc will return a signature for the given payload.
type SignFunc func(payload []byte) (signature []byte, err error)

// Encode will return a signed, URL-safe string representation of the token.
// If signFn is nil, the signature will be omitted.
func (t Token) Encode(signFn SignFunc) (string, error) {
	var b []byte

	enc := b64Encoding
	switch t.Version {
	case 0:
		return t.ID.String(), nil
	case 1:
		if t.Type != TypeSession {
			return "", validation.NewFieldError("Type", "version 1 only supports session tokens")
		}
		b = make([]byte, 17)
		b[0] = 'S'
		copy(b[1:], t.ID[:])
		enc = base64.URLEncoding
	case 2:
		b = make([]byte, 27)
		b[0] = 'V' // versioned header format
		b[1] = 2
		b[2] = byte(t.Type)
		copy(b[3:], t.ID[:])
		binary.BigEndian.PutUint64(b[19:], uint64(t.CreatedAt.Unix()))
	case 3:
		b = make([]byte, 19)
		b[0] = 'V' // versioned header format
		b[1] = 3
		b[2] = byte(t.Type)
		copy(b[3:], t.ID[:])
	default:
		return "", validation.NewFieldError("Type", "unsupported version")
	}

	if signFn != nil {
		sig, err := signFn(b)
		if err != nil {
			return "", err
		}
		b = append(b, sig...)
	}
	return enc.EncodeToString(b), nil
}
