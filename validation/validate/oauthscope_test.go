package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuthScope(t *testing.T) {

	good := func(s string, req ...string) {
		t.Helper()
		assert.NoError(t, OAuthScope("test", s, req...))
	}
	bad := func(s string, req ...string) {
		t.Helper()
		assert.Error(t, OAuthScope("test", s, req...))
	}

	good("openid")
	good("openid foo", "openid")
	good("openid bar baz")
	good("openid")
	good("openid", "openid")

	bad("openid foo", "openid2")
	bad("openid  foo")
	bad("\\asdf")
	bad("openid/", "openid")
}
