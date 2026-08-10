package webhook

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfymsg"
	"github.com/target/goalert/retry"
)

const testDeliveryID = "11111111-2222-4333-8444-555555555555"

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type trackingBody struct {
	io.Reader
	closed bool
}

func (b *trackingBody) Close() error {
	b.closed = true
	return nil
}

func testContext() context.Context {
	return (config.Config{}).Context(context.Background())
}

func testMessage(webURL string) notification.Test {
	return notification.Test{
		Base: nfymsg.Base{
			ID:   testDeliveryID,
			Dest: NewWebhookDest(webURL),
		},
	}
}

func testResponse(req *http.Request, status int, body io.ReadCloser) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       body,
		Request:    req,
	}
}

func TestSenderUsesInjectedHTTPClient(t *testing.T) {
	ctx := testContext()
	body := &trackingBody{Reader: strings.NewReader("ignored")}
	var called bool
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			called = true
			assert.Equal(t, "test", req.URL.Scheme)
			assert.Equal(t, testDeliveryID, req.Header.Get(idempotencyKeyHeader))
			return testResponse(req, http.StatusNoContent, body), nil
		}),
	}

	result, err := NewSender(ctx, client).SendMessage(ctx, testMessage("test://gateway.invalid/hook"))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, notification.StateSent, result.State)
	assert.True(t, called, "injected HTTP client was not called")
	assert.True(t, body.closed, "response body was not closed")
}

func TestSenderHTTPStatusAndResponseBody(t *testing.T) {
	const (
		webURL          = "https://gateway.invalid/v1/goalert/contact-method/opaque-secret-token?route=secret-query"
		responseContent = "sensitive-response-content"
	)
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{name: "200", status: http.StatusOK},
		{name: "204", status: http.StatusNoContent},
		{name: "302", status: http.StatusFound, wantErr: true},
		{name: "400", status: http.StatusBadRequest, wantErr: true},
		{name: "429", status: http.StatusTooManyRequests, wantErr: true},
		{name: "500", status: http.StatusInternalServerError, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := testContext()
			body := &trackingBody{Reader: strings.NewReader(responseContent)}
			client := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return testResponse(req, test.status, body), nil
				}),
			}

			result, err := NewSender(ctx, client).SendMessage(ctx, testMessage(webURL))
			assert.True(t, body.closed, "response body was not closed")
			if !test.wantErr {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, notification.StateSent, result.State)
				return
			}

			require.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), strconv.Itoa(test.status))
			assert.NotContains(t, err.Error(), webURL)
			assert.NotContains(t, err.Error(), "opaque-secret-token")
			assert.NotContains(t, err.Error(), "secret-query")
			assert.NotContains(t, err.Error(), responseContent)
		})
	}
}

func TestSenderSanitizesConnectionError(t *testing.T) {
	const (
		webURL        = "https://gateway.invalid/v1/goalert/contact-method/opaque-secret-token?route=secret-query"
		requestSecret = "sensitive-request-content"
	)
	ctx := testContext()
	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("dial failed for " + webURL + ": sensitive-response-content")
		}),
	}
	msg := notification.Verification{
		Base: nfymsg.Base{
			ID:   testDeliveryID,
			Dest: NewWebhookDest(webURL),
		},
		Code: requestSecret,
	}

	result, err := NewSender(ctx, client).SendMessage(ctx, msg)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "webhook request failed", err.Error())
	assert.True(t, retry.IsTemporaryError(err), "connection errors must retain their existing retry classification")
	assert.NotContains(t, err.Error(), "opaque-secret-token")
	assert.NotContains(t, err.Error(), "secret-query")
	assert.NotContains(t, err.Error(), requestSecret)
	assert.NotContains(t, err.Error(), "sensitive-response-content")
}

func TestSenderRejectsMissingDeliveryIdentity(t *testing.T) {
	const (
		webURL        = "https://gateway.invalid/v1/goalert/contact-method/opaque-secret-token?route=secret-query"
		requestSecret = "sensitive-request-content"
	)
	ctx := testContext()
	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			t.Fatal("HTTP client must not be called without a delivery identity")
			return nil, nil
		}),
	}
	msg := notification.Verification{
		Base: nfymsg.Base{
			Dest: NewWebhookDest(webURL),
		},
		Code: requestSecret,
	}

	result, err := NewSender(ctx, client).SendMessage(ctx, msg)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "webhook delivery identity is required", err.Error())
	assert.NotContains(t, err.Error(), webURL)
	assert.NotContains(t, err.Error(), "opaque-secret-token")
	assert.NotContains(t, err.Error(), "secret-query")
	assert.NotContains(t, err.Error(), requestSecret)
}

func TestSenderContextFailure(t *testing.T) {
	const webURL = "https://gateway.invalid/v1/goalert/contact-method/opaque-secret-token?route=secret-query"
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			<-req.Context().Done()
			return nil, req.Context().Err()
		}),
	}

	t.Run("canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(testContext())
		cancel()

		result, err := NewSender(ctx, client).SendMessage(ctx, testMessage(webURL))
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, context.Canceled)
		assert.False(t, retry.IsTemporaryError(err), "cancellation must retain its existing retry classification")
		assert.NotContains(t, err.Error(), "opaque-secret-token")
		assert.NotContains(t, err.Error(), "secret-query")
	})

	t.Run("timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(testContext(), 10*time.Millisecond)
		defer cancel()

		result, err := NewSender(ctx, client).SendMessage(ctx, testMessage(webURL))
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.False(t, retry.IsTemporaryError(err), "deadline failures must retain their existing retry classification")
		assert.NotContains(t, err.Error(), "opaque-secret-token")
		assert.NotContains(t, err.Error(), "secret-query")
	})
}

func TestSenderSanitizesRequestConstructionError(t *testing.T) {
	const webURL = "https://gateway.invalid/opaque-secret-token\n?route=secret-query"
	ctx := testContext()
	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			t.Fatal("HTTP client must not be called for an invalid request URL")
			return nil, nil
		}),
	}

	result, err := NewSender(ctx, client).SendMessage(ctx, testMessage(webURL))
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "webhook request could not be created", err.Error())
	assert.NotContains(t, err.Error(), "opaque-secret-token")
	assert.NotContains(t, err.Error(), "secret-query")
}
