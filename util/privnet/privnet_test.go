package privnet

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsPrivateAddr(t *testing.T) {
	check := func(private bool, addr string) {
		t.Helper()
		assert.Equalf(t, private, IsPrivateAddr(netip.MustParseAddr(addr)), "IsPrivateAddr(%q)", addr)
	}

	check(false, "8.8.8.8")
	check(false, "2001:4860:4860::8888")

	check(true, "10.0.0.1")
	check(true, "172.16.0.1")
	check(true, "192.168.1.1")
	check(true, "127.0.0.1")
	check(true, "0.0.0.0")
	check(true, "169.254.169.254")
	check(true, "224.0.0.1")
	check(true, "::1")
	check(true, "::")
	check(true, "fe80::1")
	check(true, "fd00::1")
	check(true, "ff02::1")
	check(true, "::ffff:10.0.0.1") // IPv4-mapped IPv6
}

func TestRoundTripper(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	}))
	defer srv.Close()

	base := http.DefaultTransport.(*http.Transport).Clone()
	base.Proxy = nil
	client := &http.Client{Transport: RoundTripper(base)}

	// unflagged requests are allowed (test server listens on 127.0.0.1)
	req, err := http.NewRequestWithContext(context.Background(), "GET", srv.URL, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, 1, calls)

	// flagged requests are rejected at dial time
	req, err = http.NewRequestWithContext(WithBlockPrivate(context.Background()), "GET", srv.URL, nil)
	require.NoError(t, err)
	_, err = client.Do(req)
	require.ErrorIs(t, err, ErrPrivateAddress)
	assert.Equal(t, 1, calls, "request should not have been sent")
}

func TestRoundTripper_Proxy(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	}))
	defer srv.Close()

	base := http.DefaultTransport.(*http.Transport).Clone()
	base.Proxy = func(*http.Request) (*url.URL, error) { return url.Parse(srv.URL) }
	client := &http.Client{Transport: RoundTripper(base)}

	// flagged request to a private destination is rejected before the proxy
	// is contacted
	req, err := http.NewRequestWithContext(WithBlockPrivate(context.Background()), "GET", "http://localhost/", nil)
	require.NoError(t, err)
	_, err = client.Do(req)
	require.True(t, errors.Is(err, ErrPrivateAddress), "expected ErrPrivateAddress, got: %v", err)
	assert.Equal(t, 0, calls, "proxy should not have been contacted")
}
