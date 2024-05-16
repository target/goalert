package authtoken

import (
	"encoding/base64"
	"encoding/binary"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/validation"
)

// A VerifyFunc will verify that the signature is valid for the given type and payload. Additionally,
// if supported, it indicates if an old-but-still-valid key (since been rotated) was used
// to generate the signature.
type VerifyFunc func(t Type, payload, signature []byte) (isValid, isOldKey bool)

func alwaysValid(Type, []byte, []byte) (bool, bool) { return true, false }

// Parse will parse a token string, optionally verifying it's signature.
// If verifyFn is nil, the signature is ignored.
func Parse(s string, verifyFn VerifyFunc) (*Token, bool, error) {
	if len(s) == 0 {
		return nil, false, validation.NewGenericError("invalid length")
	}

	if len(s) == 36 {
		// integration key type is the only one with possible length 36. Session keys, even if
		// we switched to a 128-bit signature would be a minimum of 38 base64-encoded chars.
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, false, validation.NewGenericError(err.Error())
		}
		return &Token{ID: id}, false, nil
	}
	if verifyFn == nil {
		verifyFn = alwaysValid
	}

	enc := b64Encoding
	if s[0] == 'U' { // always 'U' if it's an encoded session token (first encoded byte is 'S')
		// session tokens were/are encoded with padding enabled
		enc = base64.URLEncoding
	}

	data, err := enc.DecodeString(s)
	if err != nil {
		return nil, false, validation.NewGenericError(err.Error())
	}
	if len(data) < 2 {
		return nil, false, validation.NewGenericError("invalid length")
	}
	headerFormat := data[0]
	if headerFormat == 'S' {
		// session token (version: 1)
		if len(data) < 17 {
			return nil, false, validation.NewGenericError("invalid length")
		}
		valid, isOldKey := verifyFn(TypeSession, data[:17], data[17:])
		if !valid {
			return nil, false, validation.NewGenericError("invalid signature")
		}
		t := &Token{
			Version: 1,
			Type:    TypeSession,
		}
		copy(t.ID[:], data[1:])
		return t, isOldKey, nil
	}

	// Ensure we're using the new "versioned" token format
	if headerFormat != 'V' {
		return nil, false, validation.NewGenericError("invalid token header format")
	}

	switch data[1] {
	case 3:
		if len(data) < 19 {
			return nil, false, validation.NewGenericError("invalid length")
		}
		t := &Token{
			Version: 3,
			Type:    Type(data[2]),
		}
		valid, isOldKey := verifyFn(t.Type, data[:19], data[19:])
		if !valid {
			return nil, false, validation.NewGenericError("invalid signature")
		}
		copy(t.ID[:], data[3:])
		return t, isOldKey, nil
	case 2:
		if len(data) < 27 {
			return nil, false, validation.NewGenericError("invalid length")
		}
		t := &Token{
			Version: 2,
			Type:    Type(data[2]),
		}
		valid, isOldKey := verifyFn(t.Type, data[:27], data[27:])
		if !valid {
			return nil, false, validation.NewGenericError("invalid signature")
		}
		copy(t.ID[:], data[3:])
		t.CreatedAt = time.Unix(int64(binary.BigEndian.Uint64(data[19:])), 0)
		return t, isOldKey, nil
	}

	return nil, false, validation.NewGenericError("unsupported token version")
}
