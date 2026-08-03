package cloudwatch

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	// maxCachedCerts bounds the cache. The cert URL path is attacker-supplied,
	// so without a bound `https://sns.<region>.amazonaws.com/<random>.pem` mints
	// unlimited entries. SNS's real working set is one or two certs.
	maxCachedCerts = 32

	// maxCertBytes bounds the response read. The inbound request body limit does
	// not apply to responses we fetch.
	maxCertBytes = 16 * 1024

	certFetchTimeout = 5 * time.Second
)

// certCache maps a validated signing-cert URL to its RSA public key.
//
// Caching by URL is safe because AWS mints a new URL when it rotates a cert, so
// a stale entry becomes unreachable rather than wrong. If AWS ever reused a URL
// with new key material, verification would fail until restart.
//
// Eviction is FIFO rather than LRU: with a working set of one or two, insertion
// order is sufficient and far simpler.
type certCache struct {
	mx    sync.Mutex
	keys  map[string]*rsa.PublicKey
	order []string
}

func newCertCache() *certCache {
	return &certCache{keys: make(map[string]*rsa.PublicKey, maxCachedCerts)}
}

func (c *certCache) get(key string) (*rsa.PublicKey, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()
	pub, ok := c.keys[key]
	return pub, ok
}

func (c *certCache) put(key string, pub *rsa.PublicKey) {
	c.mx.Lock()
	defer c.mx.Unlock()

	if _, ok := c.keys[key]; ok {
		return
	}
	for len(c.order) >= maxCachedCerts {
		delete(c.keys, c.order[0])
		c.order = c.order[1:]
	}
	c.keys[key] = pub
	c.order = append(c.order, key)
}

// publicKey returns the RSA public key for the signing cert at rawURL, fetching
// and caching it if needed. rawURL is validated against the host allowlist
// before any request is made; dial maps the validated URL to the origin to
// actually request (identity in production).
func (c *certCache) publicKey(ctx context.Context, hc *http.Client, dial func(*url.URL) *url.URL, rawURL string) (*rsa.PublicKey, error) {
	u, err := checkCertURL(rawURL)
	if err != nil {
		return nil, err
	}
	key := u.String()

	// Note the lock is not held across the fetch below: holding it would
	// serialize every request behind a single AWS call.
	if pub, ok := c.get(key); ok {
		return pub, nil
	}

	ctx, cancel := context.WithTimeout(ctx, certFetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dial(u).String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cloudwatch: build cert request: %w", err)
	}

	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cloudwatch: fetch signing certificate: %w", err)
	}
	defer resp.Body.Close()

	// A blocked redirect also lands here, since the client is configured to
	// surface 3xx rather than follow it.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cloudwatch: fetch signing certificate: %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxCertBytes))
	if err != nil {
		return nil, fmt.Errorf("cloudwatch: read signing certificate: %w", err)
	}

	// Only a successfully parsed key is cached, so a 404 or garbage body caches
	// nothing.
	pub, err := parseCertPublicKey(body)
	if err != nil {
		return nil, err
	}
	c.put(key, pub)

	return pub, nil
}
