package cloudwatch

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// snsHostRe matches an AWS SNS API hostname. The anchors are load-bearing:
// without them, `sns.us-west-2.amazonaws.com.evil.com` is accepted.
//
// The `[a-z0-9-]+` region segment deliberately forbids dots, so GovCloud
// (sns.us-gov-west-1.amazonaws.com) matches while ISO regions (c2s.ic.gov,
// sc2s.sgov.gov) do not.
var snsHostRe = regexp.MustCompile(`^sns\.[a-z0-9-]+\.amazonaws\.com(\.cn)?$`)

// certPathRe matches a single *.pem path segment. Restricting the path bounds
// how many distinct cache keys an attacker-supplied URL can mint.
var certPathRe = regexp.MustCompile(`^/[A-Za-z0-9._-]+\.pem$`)

var errNotAllowed = errors.New("cloudwatch: URL host is not an AWS SNS endpoint")

// checkFetchURL validates raw as an AWS SNS endpoint and returns the parsed URL.
//
// Callers MUST build their request from the returned *url.URL and never re-parse
// raw: validating one string while dialing another is the classic bypass.
func checkFetchURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errNotAllowed, err)
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return nil, fmt.Errorf("%w: scheme %q", errNotAllowed, u.Scheme)
	}
	// Reject userinfo: `https://sns.us-west-2.amazonaws.com@evil.com/` has a host
	// of evil.com, and we never want to leak credentials outbound either.
	if u.User != nil {
		return nil, fmt.Errorf("%w: URL contains userinfo", errNotAllowed)
	}
	// AWS never sends an explicit port. Rejecting it keeps the host check to a
	// single auditable line and avoids hand-rolled port stripping.
	if u.Port() != "" {
		return nil, fmt.Errorf("%w: explicit port not allowed", errNotAllowed)
	}

	// Hostname() strips the port and IPv6 brackets; u.Host does not. Lowercase
	// before matching -- DNS is case-insensitive, so this is correctness.
	host := strings.ToLower(u.Hostname())
	if !snsHostRe.MatchString(host) {
		return nil, fmt.Errorf("%w: host %q", errNotAllowed, host)
	}

	return u, nil
}

// checkCertURL validates a SigningCertURL. Beyond checkFetchURL it requires a
// bare `/<name>.pem` path with no query or fragment.
func checkCertURL(raw string) (*url.URL, error) {
	u, err := checkFetchURL(raw)
	if err != nil {
		return nil, err
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("%w: cert URL must not have a query or fragment", errNotAllowed)
	}
	if !certPathRe.MatchString(u.Path) {
		return nil, fmt.Errorf("%w: cert path %q is not a single .pem segment", errNotAllowed, u.Path)
	}

	return u, nil
}

// checkSubscribeURL validates a SubscribeURL. Unlike a cert URL it legitimately
// carries ?Action=ConfirmSubscription&TopicArn=...&Token=...
func checkSubscribeURL(raw string) (*url.URL, error) { return checkFetchURL(raw) }
