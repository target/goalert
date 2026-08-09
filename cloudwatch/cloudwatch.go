// Package cloudwatch implements ingress for AWS CloudWatch alarms delivered over
// Amazon SNS.
//
// Unlike GoAlert's other ingress handlers this one performs the SNS subscription
// handshake (fetching SubscribeURL) and verifies the RSA message signature, so it
// can be subscribed to an SNS topic directly with no intermediate forwarder.
//
// # Trust model
//
// A valid SNS signature proves only that some SNS topic in some AWS account
// signed the message. It is NOT authorization -- the integration key is. The
// topic is not pinned to a key, so any AWS account holding the key token can
// deliver to it.
//
// The trust anchor for the signing certificate is TLS plus the host allowlist,
// not the certificate's contents: we cannot chain-validate it without pinning an
// AWS CA. That is why the allowlist in allowlist.go and the redirect blocking in
// this file are the load-bearing controls.
package cloudwatch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util/calllimiter"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
)

const (
	// maxBodyBytes bounds the request body independently of the global
	// maxBodySizeMiddleware, which is disabled when MaxReqBodyBytes is 0. SNS
	// caps messages at 256KB.
	maxBodyBytes = 256 * 1024

	confirmTimeout  = 10 * time.Second
	maxConfirmBytes = 8 * 1024
)

// Config configures a Handler.
type Config struct {
	AlertStore          *alert.Store
	IntegrationKeyStore *integrationkey.Store

	// Client is used to fetch signing certificates and confirm subscriptions. If
	// nil, a default is used. A supplied client MUST NOT follow redirects.
	Client *http.Client

	// BaseURL overrides the origin used for outbound requests. Testing only; the
	// host allowlist still runs against the message-supplied URL first.
	BaseURL string
}

// Handler serves the CloudWatch/SNS ingress endpoint.
type Handler struct {
	cfg   Config
	base  *url.URL
	hc    *http.Client
	certs *certCache
}

// NewHandler returns a Handler for the given config.
func NewHandler(cfg Config) (*Handler, error) {
	h := &Handler{cfg: cfg, hc: cfg.Client, certs: newCertCache()}
	if h.hc == nil {
		h.hc = defaultClient()
	}
	if cfg.BaseURL != "" {
		u, err := url.Parse(cfg.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("cloudwatch: parse BaseURL: %w", err)
		}
		h.base = u
	}

	return h, nil
}

// defaultClient returns a client that never follows redirects.
//
// This is not optional: the host allowlist only covers the first hop, so a
// single 302 from an allowlisted host would otherwise reach an arbitrary
// address, including the EC2 instance metadata endpoint.
func defaultClient() *http.Client {
	return &http.Client{
		Transport: calllimiter.RoundTripper(http.DefaultTransport),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// dialURL maps an already-validated URL to the origin to actually request. In
// production (BaseURL empty) it returns u unchanged.
func (h *Handler) dialURL(u *url.URL) *url.URL {
	if h.base == nil {
		return u
	}

	// Scheme is replaced along with the host so a plain-HTTP test server works
	// against an allowlist that requires HTTPS. Path and query are preserved.
	out := *u
	out.Scheme = h.base.Scheme
	out.Host = h.base.Host
	out.User = nil
	out.Path = h.base.Path + u.Path

	return &out
}

// ServeIncoming handles an SNS message.
//
// Invariant: no outbound request other than the signing-certificate fetch, and
// no alert write, happens before the signature is verified. The cert fetch
// necessarily precedes verification, which is why checkCertURL must be airtight.
func (h *Handler) ServeIncoming(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := permission.LimitCheckAny(ctx, permission.Service)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	// MaxBytesReader rather than io.LimitReader: errutil.HTTPError maps
	// *http.MaxBytesError to a clean 413, whereas a silent truncation would
	// surface as a confusing 400 or 403.
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	data, err := io.ReadAll(r.Body)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	// Deliberately not gated on Content-Type: SNS always sends text/plain and
	// offers no way to change it. Gating is the bug that makes the generic
	// endpoint unusable for SNS.
	var e envelope
	if err := json.Unmarshal(data, &e); err != nil {
		log.Debugf(ctx, "cloudwatch: bad request body: %v", err)
		clientError(w, http.StatusBadRequest)
		return
	}

	ctx = log.WithFields(ctx, log.Fields{
		"SNSType":      e.Type,
		"TopicArn":     e.TopicARN,
		"SNSMessageID": e.MessageID,
	})

	// Must precede verification: the signed field list depends on the type.
	if _, ok := signedFields[e.Type]; !ok {
		log.Debugf(ctx, "cloudwatch: unknown message type %q", e.Type)
		clientError(w, http.StatusBadRequest)
		return
	}
	if e.SigningCertURL == "" || e.Signature == "" || e.SignatureVersion == "" || e.MessageID == "" || e.TopicARN == "" {
		log.Debugf(ctx, "cloudwatch: envelope is missing required fields")
		clientError(w, http.StatusBadRequest)
		return
	}

	u, err := checkCertURL(e.SigningCertURL)
	if err != nil {
		// Info level: a rejected cert URL is a security-relevant event, but it is
		// caller-driven and must not fail the smoke harness.
		log.Logf(ctx, "cloudwatch: rejected signing cert URL: %v", err)
		clientError(w, http.StatusForbidden)
		return
	}

	pub, err := h.certs.publicKey(ctx, h.hc, h.dialURL, u.String())
	if err != nil {
		// Transient: 500 so SNS retries rather than dropping the alarm.
		log.Log(ctx, fmt.Errorf("cloudwatch: signing certificate: %w", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := verifyMessage(time.Now(), pub, &e); err != nil {
		log.Logf(ctx, "cloudwatch: %v", err)
		clientError(w, http.StatusForbidden)
		return
	}

	switch e.Type {
	case typeSubscriptionConfirmation:
		h.serveConfirmation(ctx, w, &e)
	case typeUnsubscribeConfirmation:
		// Log loudly: someone detached a live alarm feed. Deliberately do NOT
		// fetch SubscribeURL -- re-confirming would resurrect a subscription
		// somebody removed on purpose.
		log.Log(ctx, fmt.Errorf("cloudwatch: subscription removed for topic %s", e.TopicARN))
		w.WriteHeader(http.StatusOK)
	case typeNotification:
		h.serveNotification(ctx, w, &e)
	default:
		clientError(w, http.StatusBadRequest)
	}
}

// serveConfirmation completes the SNS subscription handshake. The signature has
// already been verified, and SubscribeURL is inside the signed field set, so the
// URL is known to have come from AWS.
func (h *Handler) serveConfirmation(ctx context.Context, w http.ResponseWriter, e *envelope) {
	u, err := checkSubscribeURL(e.SubscribeURL)
	if err != nil {
		// A valid signature over a non-AWS host should be impossible.
		log.Log(ctx, fmt.Errorf("cloudwatch: rejected SubscribeURL: %w", err))
		clientError(w, http.StatusBadRequest)
		return
	}

	reqCtx, cancel := context.WithTimeout(ctx, confirmTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, h.dialURL(u).String(), nil)
	if err != nil {
		log.Log(ctx, fmt.Errorf("cloudwatch: build confirmation request: %w", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp, err := h.hc.Do(req)
	if err != nil {
		log.Log(ctx, fmt.Errorf("cloudwatch: confirm subscription: %w", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// Drain and close for connection reuse. The body is never inspected and never
	// echoed anywhere: reflecting it would turn a blind SSRF into a read oracle.
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxConfirmBytes))
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 500 so SNS re-sends; the confirmation token stays valid for 3 days.
		log.Log(ctx, fmt.Errorf("cloudwatch: confirm subscription: %s", resp.Status))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// A new alarm feed attaching to a service is rare and worth recording. Note
	// the Token and full SubscribeURL are deliberately not logged: the Token can
	// confirm or cancel the subscription.
	log.Logf(ctx, "cloudwatch: confirmed subscription for topic %s", e.TopicARN)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) serveNotification(ctx context.Context, w http.ResponseWriter, e *envelope) {
	a, meta, ok := buildAlert(*e)
	if !ok {
		// INSUFFICIENT_DATA transitions are noise. 2xx so SNS does not retry.
		w.WriteHeader(http.StatusNoContent)
		return
	}
	a.ServiceID = permission.ServiceID(ctx)
	if len(meta) == 0 {
		meta = nil
	}

	var created *alert.Alert
	err := retry.DoTemporaryError(func(int) error {
		var err error
		created, _, err = h.cfg.AlertStore.CreateOrUpdateWithMeta(ctx, &a, meta)
		return err
	},
		retry.Log(ctx),
		// Deliberately shorter than genericapi's Limit(10)+1s backoff: SNS aborts
		// an HTTPS delivery at 15s, and under an alarm storm long in-handler
		// retries just convert the storm into queue-full 429s. Let SNS handle the
		// long-horizon retrying.
		retry.Limit(5),
		retry.FibBackoff(250*time.Millisecond),
	)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	// created is nil with a nil error when the status was closed and no open
	// alert held this dedup key. A stray OK is normal, so this is info, not an
	// error, and not a 500.
	if created == nil {
		log.Logf(ctx, "cloudwatch: no open alert to close")
	}

	w.WriteHeader(http.StatusNoContent)
}

// clientError writes a bare status text, never the underlying error.
func clientError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
