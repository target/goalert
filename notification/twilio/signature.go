package twilio

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"net/url"
	"sort"
)

// Signature will calculate the raw signature for a request from Twilio.
// https://www.twilio.com/docs/api/security#validating-requests
func Signature(authToken, url string, fields url.Values) []byte {
	buf := new(bytes.Buffer)
	buf.WriteString(url)

	fieldNames := make(sort.StringSlice, 0, len(fields))
	for name := range fields {
		fieldNames = append(fieldNames, name)
	}
	fieldNames.Sort()

	for _, fieldName := range fieldNames {
		buf.WriteString(fieldName + fields.Get(fieldName))
	}

	hash := hmac.New(sha1.New, []byte(authToken))
	_, _ = io.Copy(hash, buf)

	buf.Reset()
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	_, _ = enc.Write(hash.Sum(nil))
	_ = enc.Close()

	return buf.Bytes()
}
