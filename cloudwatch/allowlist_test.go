package cloudwatch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckFetchURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		ok   bool
	}{
		{name: "valid us-west-2", url: "https://sns.us-west-2.amazonaws.com/x.pem", ok: true},
		{name: "valid china partition", url: "https://sns.cn-north-1.amazonaws.com.cn/x.pem", ok: true},
		{name: "valid govcloud", url: "https://sns.us-gov-west-1.amazonaws.com/x.pem", ok: true},
		{name: "uppercase host is same DNS name", url: "https://SNS.US-WEST-2.AMAZONAWS.COM/x.pem", ok: true},

		// The case an unanchored regex would accept.
		{name: "suffix attack", url: "https://sns.us-west-2.amazonaws.com.evil.com/x.pem"},
		{name: "http scheme", url: "http://sns.us-west-2.amazonaws.com/x.pem"},
		{name: "prefix attack", url: "https://notsns.us-west-2.amazonaws.com/x.pem"},
		{name: "explicit port", url: "https://sns.us-west-2.amazonaws.com:8443/x.pem"},
		{name: "userinfo host is evil.com", url: "https://sns.us-west-2.amazonaws.com@evil.com/x.pem"},
		{name: "at sign in path", url: "https://evil.com/@sns.us-west-2.amazonaws.com/x.pem"},
		{name: "empty region segment", url: "https://sns..amazonaws.com/x.pem"},
		{name: "trailing dot", url: "https://sns.us-west-2.amazonaws.com./x.pem"},
		{name: "metadata endpoint", url: "https://169.254.169.254/latest/meta-data/"},
		{name: "percent encoded host", url: "https://%73ns.us-west-2.amazonaws.com/x.pem"},
		{name: "scheme relative", url: "//sns.us-west-2.amazonaws.com/x.pem"},
		{name: "empty", url: ""},
		{name: "dotted region", url: "https://sns.a.b.amazonaws.com/x.pem"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := checkFetchURL(tt.url)
			if !tt.ok {
				assert.Error(t, err, "expected %q to be refused", tt.url)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, u)
		})
	}
}

func TestCheckCertURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		ok   bool
	}{
		{name: "pem path", url: "https://sns.us-west-2.amazonaws.com/SimpleNotificationService-abc123.pem", ok: true},
		{name: "query string", url: "https://sns.us-west-2.amazonaws.com/x.pem?a=1"},
		{name: "fragment", url: "https://sns.us-west-2.amazonaws.com/x.pem#a"},
		{name: "nested path", url: "https://sns.us-west-2.amazonaws.com/a/b.pem"},
		{name: "not a pem", url: "https://sns.us-west-2.amazonaws.com/x.txt"},
		{name: "no extension", url: "https://sns.us-west-2.amazonaws.com/x"},
		{name: "bad host still refused", url: "https://sns.us-west-2.amazonaws.com.evil.com/x.pem"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := checkCertURL(tt.url)
			if tt.ok {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
		})
	}
}

func TestCheckSubscribeURL_AllowsQuery(t *testing.T) {
	// SubscribeURL legitimately carries the confirmation action and token.
	_, err := checkSubscribeURL("https://sns.us-west-2.amazonaws.com/?Action=ConfirmSubscription&TopicArn=arn&Token=abc")
	assert.NoError(t, err)
}
