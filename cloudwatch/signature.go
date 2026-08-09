package cloudwatch

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"time"
)

// Freshness window for the signed Timestamp. Timestamp is inside the signed
// field set but AWS does not check it for us, so without this a captured valid
// envelope replays forever.
const (
	maxMessageAge   = time.Hour
	maxMessageSkew  = 5 * time.Minute
	minCertKeyBits  = 2048
	timestampFormat = "2006-01-02T15:04:05.000Z"
)

// errBadSignature wraps every verification failure so the handler has exactly
// one branch to map to 403 and the client learns nothing about which step
// failed. Detail goes to the server log.
var errBadSignature = errors.New("cloudwatch: signature verification failed")

// verifyMessage verifies e's signature against pub and checks its freshness.
//
// Pure: no I/O. now is a parameter rather than time.Now() so the freshness
// window is table-testable.
func verifyMessage(now time.Time, pub *rsa.PublicKey, e *envelope) error {
	// Select the hash first: the canonical form could differ in a future
	// version, so an unknown version must never fall back to SHA-1.
	var hashID crypto.Hash
	switch e.SignatureVersion {
	case "1":
		// SHA-1 is dictated by AWS -- it is their default SignatureVersion, not
		// a choice we get to make here.
		hashID = crypto.SHA1
	case "2":
		hashID = crypto.SHA256
	default:
		return fmt.Errorf("%w: unsupported SignatureVersion %q", errBadSignature, e.SignatureVersion)
	}

	sig, err := base64.StdEncoding.DecodeString(e.Signature)
	if err != nil {
		return fmt.Errorf("%w: decode signature: %v", errBadSignature, err)
	}
	if len(sig) == 0 {
		return fmt.Errorf("%w: empty signature", errBadSignature)
	}

	str, err := canonicalString(e)
	if err != nil {
		return fmt.Errorf("%w: %v", errBadSignature, err)
	}

	var digest []byte
	if hashID == crypto.SHA1 {
		sum := sha1.Sum([]byte(str))
		digest = sum[:]
	} else {
		sum := sha256.Sum256([]byte(str))
		digest = sum[:]
	}

	if err := rsa.VerifyPKCS1v15(pub, hashID, digest, sig); err != nil {
		return fmt.Errorf("%w: %v", errBadSignature, err)
	}

	return checkFreshness(now, e.Timestamp)
}

func checkFreshness(now time.Time, timestamp string) error {
	ts, err := time.Parse(timestampFormat, timestamp)
	if err != nil {
		// SNS also documents plain RFC3339.
		ts, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return fmt.Errorf("%w: parse Timestamp %q", errBadSignature, timestamp)
		}
	}

	if age := now.Sub(ts); age > maxMessageAge {
		return fmt.Errorf("%w: message is %s old", errBadSignature, age)
	}
	if skew := ts.Sub(now); skew > maxMessageSkew {
		return fmt.Errorf("%w: message is %s in the future", errBadSignature, skew)
	}

	return nil
}

// parseCertPublicKey extracts the RSA public key from a PEM-encoded signing
// certificate.
//
// The certificate is deliberately not chain-validated and its expiry is not
// enforced: TLS to the allowlisted sns.<region>.amazonaws.com host is what
// authenticates these bytes, and enforcing NotAfter would only add a
// clock-skew alert-loss mode.
func parseCertPublicKey(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("cloudwatch: no PEM block in signing certificate")
	}
	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("cloudwatch: PEM block is %q, want CERTIFICATE", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cloudwatch: parse signing certificate: %w", err)
	}

	// Checked assertion: an ECDSA or Ed25519 cert would otherwise panic the
	// request goroutine.
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("cloudwatch: signing cert key type %T, want RSA", cert.PublicKey)
	}
	if pub.N.BitLen() < minCertKeyBits {
		return nil, fmt.Errorf("cloudwatch: signing cert key is %d bits, want >= %d", pub.N.BitLen(), minCertKeyBits)
	}

	return pub, nil
}
