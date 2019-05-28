package keyring

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestSignVerify(t *testing.T) {
	signKey, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	v := map[byte]ecdsa.PublicKey{
		0: signKey.PublicKey,
	}

	db := &DB{
		verificationKeys: v,
		signingKey:       signKey,
	}
	var buf bytes.Buffer
	try := func(t *testing.T) {
		sessID := uuid.NewV4()
		buf.WriteByte('S') // session IDs will be prefixed with an "S"
		buf.Write(sessID.Bytes())
		sig, err := db.Sign(buf.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		valid, old := db.Verify(buf.Bytes(), sig)
		if !valid {
			t.Fatal("validation failed")
		}
		if old {
			t.Fatal("old key used")
		}
		buf.Reset()
	}

	for i := 0; i < 100; i++ {
		// running multiple because not all signatures are the same size (encoded)
		t.Run("", try)
	}
}
