package cloudwatch

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTimestamp = "2026-07-30T12:00:00.000Z"

var testNow = time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)

// selfSignedPEM returns a PEM-encoded self-signed cert for pub. Only the public
// key is ever read back, so no chain is needed.
func selfSignedPEM(t *testing.T, key *rsa.PrivateKey) []byte {
	t.Helper()

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "sns.amazonaws.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

// signEnvelope signs e in place using the given version ("1" for SHA-1, AWS's
// default; "2" for SHA-256).
func signEnvelope(t *testing.T, key *rsa.PrivateKey, e *envelope, version string) {
	t.Helper()

	e.SignatureVersion = version
	str, err := canonicalString(e)
	require.NoError(t, err)

	var (
		digest []byte
		hashID crypto.Hash
	)
	if version == "2" {
		sum := sha256.Sum256([]byte(str))
		digest, hashID = sum[:], crypto.SHA256
	} else {
		sum := sha1.Sum([]byte(str))
		digest, hashID = sum[:], crypto.SHA1
	}

	sig, err := rsa.SignPKCS1v15(rand.Reader, key, hashID, digest)
	require.NoError(t, err)
	e.Signature = base64.StdEncoding.EncodeToString(sig)
}

func testEnvelope() envelope {
	return envelope{
		Type:      typeNotification,
		MessageID: "mid",
		TopicARN:  "arn:aws:sns:us-west-2:123456789012:PagerDuty-Data",
		Message:   `{"AlarmName":"x","NewStateValue":"ALARM"}`,
		Timestamp: testTimestamp,
	}
}

func TestVerifyMessage(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	other, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	t.Run("v1 sha1 round trip", func(t *testing.T) {
		e := testEnvelope()
		signEnvelope(t, key, &e, "1")
		assert.NoError(t, verifyMessage(testNow, &key.PublicKey, &e))
	})

	t.Run("v2 sha256 round trip", func(t *testing.T) {
		e := testEnvelope()
		signEnvelope(t, key, &e, "2")
		assert.NoError(t, verifyMessage(testNow, &key.PublicKey, &e))
	})

	t.Run("subscription confirmation round trip", func(t *testing.T) {
		e := testEnvelope()
		e.Type = typeSubscriptionConfirmation
		e.SubscribeURL = "https://sns.us-west-2.amazonaws.com/?Action=ConfirmSubscription&Token=tok"
		e.Token = "tok"
		signEnvelope(t, key, &e, "1")
		assert.NoError(t, verifyMessage(testNow, &key.PublicKey, &e))
	})

	t.Run("wrong key rejected", func(t *testing.T) {
		e := testEnvelope()
		signEnvelope(t, key, &e, "1")
		assert.ErrorIs(t, verifyMessage(testNow, &other.PublicKey, &e), errBadSignature)
	})

	t.Run("tampered message rejected", func(t *testing.T) {
		e := testEnvelope()
		signEnvelope(t, key, &e, "1")
		e.Message = `{"AlarmName":"tampered","NewStateValue":"ALARM"}`
		assert.ErrorIs(t, verifyMessage(testNow, &key.PublicKey, &e), errBadSignature)
	})

	t.Run("empty subject cannot forge absent subject", func(t *testing.T) {
		e := testEnvelope()
		signEnvelope(t, key, &e, "1") // signed with Subject absent
		e.Subject = strPtr("")        // now present-but-empty
		assert.ErrorIs(t, verifyMessage(testNow, &key.PublicKey, &e), errBadSignature)
	})

	t.Run("unknown signature version rejected", func(t *testing.T) {
		e := testEnvelope()
		signEnvelope(t, key, &e, "1")
		// Must not silently fall back to SHA-1: a future version could change the
		// canonical form.
		e.SignatureVersion = "3"
		assert.ErrorIs(t, verifyMessage(testNow, &key.PublicKey, &e), errBadSignature)
	})

	t.Run("empty signature version rejected", func(t *testing.T) {
		e := testEnvelope()
		signEnvelope(t, key, &e, "1")
		e.SignatureVersion = ""
		assert.ErrorIs(t, verifyMessage(testNow, &key.PublicKey, &e), errBadSignature)
	})

	t.Run("non base64 signature rejected", func(t *testing.T) {
		e := testEnvelope()
		signEnvelope(t, key, &e, "1")
		e.Signature = "!!!not base64!!!"
		assert.ErrorIs(t, verifyMessage(testNow, &key.PublicKey, &e), errBadSignature)
	})

	t.Run("empty signature rejected", func(t *testing.T) {
		e := testEnvelope()
		signEnvelope(t, key, &e, "1")
		e.Signature = ""
		assert.ErrorIs(t, verifyMessage(testNow, &key.PublicKey, &e), errBadSignature)
	})

	t.Run("unknown type rejected", func(t *testing.T) {
		e := testEnvelope()
		signEnvelope(t, key, &e, "1")
		e.Type = "Bogus"
		assert.ErrorIs(t, verifyMessage(testNow, &key.PublicKey, &e), errBadSignature)
	})
}

// Timestamp is inside the signed field set, so without a freshness window a
// captured valid envelope replays forever.
func TestVerifyMessage_Freshness(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tests := []struct {
		name string
		now  time.Time
		ok   bool
	}{
		{name: "exactly now", now: testNow, ok: true},
		{name: "one minute old", now: testNow.Add(time.Minute), ok: true},
		{name: "59 minutes old", now: testNow.Add(59 * time.Minute), ok: true},
		{name: "two hours old", now: testNow.Add(2 * time.Hour)},
		{name: "one minute in future", now: testNow.Add(-time.Minute), ok: true},
		{name: "ten minutes in future", now: testNow.Add(-10 * time.Minute)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := testEnvelope()
			signEnvelope(t, key, &e, "1")
			err := verifyMessage(tt.now, &key.PublicKey, &e)
			if tt.ok {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, errBadSignature)
		})
	}
}

func TestVerifyMessage_UnparsableTimestamp(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	e := testEnvelope()
	e.Timestamp = "not a timestamp"
	signEnvelope(t, key, &e, "1")
	assert.ErrorIs(t, verifyMessage(testNow, &key.PublicKey, &e), errBadSignature)
}

func TestVerifyMessage_RFC3339Timestamp(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	e := testEnvelope()
	e.Timestamp = "2026-07-30T12:00:00Z"
	signEnvelope(t, key, &e, "1")
	assert.NoError(t, verifyMessage(testNow, &key.PublicKey, &e))
}

func TestParseCertPublicKey(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	t.Run("valid rsa cert", func(t *testing.T) {
		pub, err := parseCertPublicKey(selfSignedPEM(t, key))
		require.NoError(t, err)
		assert.Equal(t, key.N, pub.N)
	})

	t.Run("no pem block", func(t *testing.T) {
		_, err := parseCertPublicKey([]byte("not a pem file"))
		assert.Error(t, err)
	})

	t.Run("empty input", func(t *testing.T) {
		_, err := parseCertPublicKey(nil)
		assert.Error(t, err)
	})

	t.Run("wrong pem block type", func(t *testing.T) {
		block := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("x")})
		_, err := parseCertPublicKey(block)
		assert.Error(t, err)
	})

	t.Run("garbage der", func(t *testing.T) {
		block := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("garbage")})
		_, err := parseCertPublicKey(block)
		assert.Error(t, err)
	})

	// An unchecked type assertion here would panic the request goroutine.
	t.Run("non rsa key is an error not a panic", func(t *testing.T) {
		ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		tmpl := &x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject:      pkix.Name{CommonName: "sns.amazonaws.com"},
			NotBefore:    time.Now().Add(-time.Hour),
			NotAfter:     time.Now().Add(time.Hour),
		}
		der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &ecKey.PublicKey, ecKey)
		require.NoError(t, err)

		_, err = parseCertPublicKey(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
		assert.ErrorContains(t, err, "want RSA")
	})

	t.Run("undersized key rejected", func(t *testing.T) {
		small, err := rsa.GenerateKey(rand.Reader, 1024)
		require.NoError(t, err)
		_, err = parseCertPublicKey(selfSignedPEM(t, small))
		assert.ErrorContains(t, err, "bits")
	})
}
