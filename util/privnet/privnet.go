// Package privnet provides an http.RoundTripper that can block outbound
// connections to private, loopback, and link-local addresses on a
// per-request basis.
//
// Requests opt in by carrying a context flag set with WithBlockPrivate. For
// those requests, the destination is checked at connection time, using the
// IP address actually being dialed, so DNS names and redirects that resolve
// to internal addresses are covered.
//
// When a request is routed through an HTTP proxy, the proxy performs the DNS
// lookup and connection, so only a best-effort pre-resolution check is
// possible. In that case, destination policy must be enforced at the proxy.
package privnet

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"syscall"
	"time"
)

// ErrPrivateAddress is returned when a request destination resolves to a
// private, loopback, or link-local address and blocking is enabled.
var ErrPrivateAddress = errors.New("destination resolves to a private, loopback, or link-local address")

type contextKey struct{}

// WithBlockPrivate returns a context that causes requests made with it to be
// rejected if the destination is a private, loopback, or link-local address.
func WithBlockPrivate(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey{}, true)
}

func blockPrivate(ctx context.Context) bool {
	v, _ := ctx.Value(contextKey{}).(bool)
	return v
}

// IsPrivateAddr reports whether addr is anything other than a globally
// routable unicast address (private, loopback, link-local, multicast, etc.).
func IsPrivateAddr(addr netip.Addr) bool {
	addr = addr.Unmap()
	return addr.IsPrivate() || !addr.IsGlobalUnicast()
}

// control is invoked for every connection attempt on the guarded transport,
// after DNS resolution, with the actual address being connected to.
func control(_ context.Context, _, address string, _ syscall.RawConn) error {
	ap, err := netip.ParseAddrPort(address)
	if err != nil {
		return fmt.Errorf("parse dial address %q: %w", address, err)
	}
	if IsPrivateAddr(ap.Addr()) {
		return ErrPrivateAddress
	}

	return nil
}

type roundTripper struct {
	plain   *http.Transport
	guarded *http.Transport
}

// RoundTripper wraps base so that requests flagged with WithBlockPrivate are
// sent over a separate, guarded copy of the transport that rejects private
// destinations at dial time. The guarded copy keeps its own connection pool
// so a connection established without the check is never reused by a
// flagged request. Unflagged requests use base unchanged.
func RoundTripper(base *http.Transport) http.RoundTripper {
	guarded := base.Clone()
	guarded.DialContext = (&net.Dialer{
		Timeout:        30 * time.Second,
		KeepAlive:      30 * time.Second,
		ControlContext: control,
	}).DialContext

	return &roundTripper{plain: base, guarded: guarded}
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !blockPrivate(req.Context()) {
		return rt.plain.RoundTrip(req)
	}

	if rt.guarded.Proxy != nil {
		proxyURL, err := rt.guarded.Proxy(req)
		if err != nil {
			return nil, err
		}
		if proxyURL != nil {
			// The proxy will resolve and connect on our behalf, so the dial
			// check applies to the proxy address. Do a best-effort check of
			// the destination hostname here.
			err = checkHost(req.Context(), req.URL.Hostname())
			if err != nil {
				return nil, err
			}
		}
	}

	return rt.guarded.RoundTrip(req)
}

func checkHost(ctx context.Context, host string) error {
	addrs, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return fmt.Errorf("resolve %q: %w", host, err)
	}
	for _, addr := range addrs {
		if IsPrivateAddr(addr) {
			return ErrPrivateAddress
		}
	}

	return nil
}
