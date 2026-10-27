package oidc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/auth"
	"github.com/target/goalert/auth/nonce"
	"github.com/target/goalert/config"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/util/log"
)

// hmacKeyring is a minimal in-memory keyring, only Sign and Verify are used by the provider.
type hmacKeyring struct{ keyring.Keyring }

func (hmacKeyring) Sign(p []byte) ([]byte, error) {
	m := hmac.New(sha256.New, []byte("test-key"))
	m.Write(p)
	return m.Sum(nil), nil
}

func (k hmacKeyring) Verify(p, sig []byte) (bool, bool) {
	exp, _ := k.Sign(p)
	return hmac.Equal(exp, sig), false
}

// fakeDriver backs a nonce.Store without a database; every statement affects exactly one row.
type fakeDriver struct{}

type fakeConn struct{}

type fakeStmt struct{}

func (fakeDriver) Open(string) (driver.Conn, error)         { return fakeConn{}, nil }
func (fakeConn) Prepare(string) (driver.Stmt, error)        { return fakeStmt{}, nil }
func (fakeConn) Close() error                               { return nil }
func (fakeConn) Begin() (driver.Tx, error)                  { return nil, driver.ErrSkip }
func (fakeStmt) Close() error                               { return nil }
func (fakeStmt) NumInput() int                              { return -1 }
func (fakeStmt) Exec([]driver.Value) (driver.Result, error) { return driver.RowsAffected(1), nil }
func (fakeStmt) Query([]driver.Value) (driver.Rows, error)  { return nil, driver.ErrSkip }

func init() { sql.Register("oidc-pkce-fake", fakeDriver{}) }

type pkceEnv struct {
	idp      *mockoidc.MockOIDC
	provider *Provider
	ctx      context.Context
	route    func(rel string) *auth.RouteInfo
}

func newPKCEEnv(t *testing.T, enablePKCE bool) *pkceEnv {
	t.Helper()

	idp, err := mockoidc.Run()
	require.NoError(t, err)
	t.Cleanup(func() { _ = idp.Shutdown() })

	db, err := sql.Open("oidc-pkce-fake", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	store, err := nonce.NewStore(context.Background(), log.NewLogger(), db)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Shutdown(context.Background()) })

	p, err := NewProvider(context.Background(), Config{Keyring: hmacKeyring{}, NonceStore: store})
	require.NoError(t, err)

	var cfg config.Config
	cfg.OIDC.Enable = true
	cfg.OIDC.IssuerURL = idp.Issuer()
	cfg.OIDC.ClientID = idp.ClientID
	cfg.OIDC.ClientSecret = idp.ClientSecret
	cfg.OIDC.EnablePKCE = enablePKCE

	const base = "http://goalert.test/api/v2/identity/providers/oidc"
	return &pkceEnv{
		idp:      idp,
		provider: p,
		ctx:      cfg.Context(context.Background()),
		route: func(rel string) *auth.RouteInfo {
			return &auth.RouteInfo{RelativePath: rel, CurrentURL: base}
		},
	}
}

// login starts a login and returns the authorization URL and the cookies set for the browser.
func (e *pkceEnv) login(t *testing.T) (*url.URL, map[string]*http.Cookie) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "http://goalert.test/api/v2/identity/providers/oidc", nil).WithContext(e.ctx)
	rec := httptest.NewRecorder()
	_, err := e.provider.ExtractIdentity(e.route("/"), rec, req)
	redirect, ok := err.(auth.RedirectURL)
	require.True(t, ok, "expected a redirect, got %v", err)

	u, err := url.Parse(string(redirect))
	require.NoError(t, err)

	cookies := make(map[string]*http.Cookie)
	for _, c := range rec.Result().Cookies() {
		cookies[c.Name] = c
	}
	return u, cookies
}

// authorize follows the authorization URL at the IdP and returns the callback query.
func (e *pkceEnv) authorize(t *testing.T, u *url.URL) url.Values {
	t.Helper()

	e.idp.QueueCode("test-code")
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Get(u.String())
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusFound, resp.StatusCode)

	loc, err := url.Parse(resp.Header.Get("Location"))
	require.NoError(t, err)
	return loc.Query()
}

func (e *pkceEnv) callback(t *testing.T, q url.Values, cookies ...*http.Cookie) (*auth.Identity, error) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "http://goalert.test/api/v2/identity/providers/oidc/callback?"+q.Encode(), nil).WithContext(e.ctx)
	for _, c := range cookies {
		if c != nil {
			req.AddCookie(&http.Cookie{Name: c.Name, Value: c.Value})
		}
	}
	return e.provider.ExtractIdentity(e.route("/callback"), httptest.NewRecorder(), req)
}

func TestProvider_PKCE(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		e := newPKCEEnv(t, true)
		u, cookies := e.login(t)

		assert.Equal(t, "S256", u.Query().Get("code_challenge_method"))
		assert.NotEmpty(t, u.Query().Get("code_challenge"))
		verifier := cookies[pkceVerifierCookieName]
		require.NotNil(t, verifier, "expected a verifier cookie")
		assert.True(t, verifier.HttpOnly)
		assert.NotContains(t, u.String(), verifier.Value, "the verifier must not be sent to the IdP at login")
		challenge, err := mockoidc.GenerateCodeChallenge(mockoidc.CodeChallengeMethodS256, verifier.Value)
		require.NoError(t, err)
		assert.Equal(t, challenge, u.Query().Get("code_challenge"))

		q := e.authorize(t, u)

		id, err := e.callback(t, q, cookies[nonceCookieName], verifier)
		require.NoError(t, err)
		assert.NotEmpty(t, id.SubjectID)
	})

	t.Run("enabled-missing-verifier", func(t *testing.T) {
		e := newPKCEEnv(t, true)
		u, cookies := e.login(t)
		q := e.authorize(t, u)

		_, err := e.callback(t, q, cookies[nonceCookieName])
		assert.ErrorAs(t, err, new(auth.Error))
	})

	t.Run("enabled-wrong-verifier", func(t *testing.T) {
		e := newPKCEEnv(t, true)
		u, cookies := e.login(t)
		q := e.authorize(t, u)

		// The IdP must reject a verifier that does not match the challenge.
		wrong := &http.Cookie{Name: pkceVerifierCookieName, Value: "A-different-verifier-that-is-long-enough-for-the-spec-0123456789"}
		_, err := e.callback(t, q, cookies[nonceCookieName], wrong)
		assert.ErrorAs(t, err, new(auth.Error))
	})

	t.Run("disabled", func(t *testing.T) {
		e := newPKCEEnv(t, false)
		u, cookies := e.login(t)

		assert.False(t, u.Query().Has("code_challenge"))
		assert.False(t, u.Query().Has("code_challenge_method"))
		assert.NotContains(t, cookies, pkceVerifierCookieName)

		q := e.authorize(t, u)
		id, err := e.callback(t, q, cookies[nonceCookieName])
		require.NoError(t, err)
		assert.NotEmpty(t, id.SubjectID)
	})
}
