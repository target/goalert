package twilio

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignature(t *testing.T) {

	// From twilio docs
	//  https://www.twilio.com/docs/api/security#validating-requests

	const (
		reqURL    = "https://mycompany.com/myapp.php?foo=1&bar=2"
		authToken = "12345"

		// Twilio's example code seems to be incorrect (includes an extra `=`)
		// so this is different than the test example.
		expectedSignature = "GvWf1cFY/Q7PnoempGyD5oXAezc="
	)

	v := make(url.Values)
	v.Set("Digits", "1234")
	v.Set("To", "+18005551212")
	v.Set("From", "+14158675310")
	v.Set("Caller", "+14158675310")
	v.Set("CallSid", "CA1234567890ABCDE")

	sig := Signature(authToken, reqURL, v)
	assert.Equal(t, expectedSignature, string(sig))
}
