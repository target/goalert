package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Signature generates a signature for a Slack request.
func Signature(signingSecret string, ts time.Time, data []byte) string {
	h := hmac.New(sha256.New, []byte(signingSecret))
	_, err := fmt.Fprintf(h, "v0:%d:%s", ts.Unix(), data)
	if err != nil {
		panic(err)
	}

	return "v0=" + hex.EncodeToString(h.Sum(nil))
}
