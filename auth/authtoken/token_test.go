package authtoken

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToken_Version0(t *testing.T) {
	tok := &Token{
		ID: uuid.UUID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
	}
	s, err := tok.Encode(nil)
	assert.NoError(t, err)
	// v0 integration keys are just a hex-encoded UUID
	assert.Equal(t, "01020304-0506-0708-090a-0b0c0d0e0f10", s)

	parsed, isOld, err := Parse(s, nil)
	assert.NoError(t, err)
	assert.False(t, isOld)
	assert.EqualValues(t, tok, parsed)
}

func TestToken_Version1(t *testing.T) {
	t.Run("encoding/decoding", func(t *testing.T) {
		tok := &Token{
			Version: 1,
			ID:      uuid.MustParse("a592fe16-edb5-45bb-a7d2-d109fae252bc"),
			Type:    TypeSession,
		}
		s, err := tok.Encode(func(b []byte) ([]byte, error) { return []byte("sig"), nil })
		assert.NoError(t, err)
		// v1 integration keys are just a hex-encoded UUID
		dec, err := base64.URLEncoding.DecodeString(s)
		assert.NoError(t, err) // should be valid base64

		var exp bytes.Buffer
		exp.WriteByte('S')     // Session key
		exp.Write(tok.ID[:])   // ID
		exp.WriteString("sig") // Signature
		assert.Equal(t, exp.Bytes(), dec)

		parsed, isOld, err := Parse(s, func(typ Type, p, sig []byte) (bool, bool) {
			assert.Equal(t, TypeSession, typ)
			assert.Equal(t, exp.Bytes()[:exp.Len()-3], p)
			assert.Equal(t, []byte("sig"), sig)
			return true, true
		})
		assert.NoError(t, err)
		assert.True(t, isOld)
		assert.EqualValues(t, tok, parsed)
	})

	t.Run("existing key", func(t *testing.T) {
		const exampleKey = "U9obklyVC0wduWIy75nbivABDxwc-rANyqNA4CZQzhkJHuNlUCfJDPpcG6W9bEIPddqPbh-sxMS1Km87jC9yLASp3i1UWtdDu2udCzM="
		parsed, isOld, err := Parse(exampleKey, func(Type, []byte, []byte) (bool, bool) { return true, true })
		assert.NoError(t, err)
		assert.True(t, isOld)
		assert.EqualValues(t, &Token{
			Type:    TypeSession,
			Version: 1,
			ID:      uuid.MustParse("da1b925c-950b-4c1d-b962-32ef99db8af0"),
		}, parsed)
	})
}

func TestToken_Version2(t *testing.T) {
	tok := &Token{
		Version:   2,
		Type:      TypeCalSub,
		ID:        uuid.UUID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		CreatedAt: time.Unix(1337, 0),
	}
	s, err := tok.Encode(func(b []byte) ([]byte, error) { return []byte("sig"), nil })
	assert.NoError(t, err)
	// v2 integration keys are just a hex-encoded UUID
	dec, err := b64Encoding.DecodeString(s)
	assert.NoError(t, err) // should be valid base64

	var exp bytes.Buffer
	exp.WriteByte('V')                                                     // Versioned header flag
	exp.WriteByte(2)                                                       // version
	exp.WriteByte(byte(TypeCalSub))                                        // type
	exp.Write(tok.ID[:])                                                   // ID
	_ = binary.Write(&exp, binary.BigEndian, uint64(tok.CreatedAt.Unix())) // CreatedAt
	exp.WriteString("sig")                                                 // Signature
	assert.Equal(t, exp.Bytes(), dec)

	parsed, isOld, err := Parse(s, func(typ Type, p, sig []byte) (bool, bool) {
		assert.Equal(t, TypeCalSub, typ)
		assert.Equal(t, exp.Bytes()[:exp.Len()-3], p)
		assert.Equal(t, sig, []byte("sig"))
		return true, true
	})
	assert.NoError(t, err)
	assert.True(t, isOld)
	assert.EqualValues(t, tok, parsed)
}
