package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/auth/authtoken"
	"github.com/target/goalert/config"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// CookieName is the name of the auth session cookie.
const (
	CookieName   = "goalert_session.2"
	v1CookieName = "goalert_session"
)

type registeredProvider struct {
	// ID is the unique identifier of the provider.
	ID string

	// URL is the location of the form action (POST).
	URL string

	ProviderInfo
}

// Handler will serve authentication requests for registered identity providers.
type Handler struct {
	providers map[string]IdentityProvider
	cfg       HandlerConfig

	db         *sql.DB
	userLookup *sql.Stmt
	addSubject *sql.Stmt
	updateUA   *sql.Stmt
	updateUser *sql.Stmt

	startSession *sql.Stmt
	fetchSession *sql.Stmt
	endSession   *sql.Stmt

	userSessions       *sql.Stmt
	endSessionUser     *sql.Stmt
	endAllSessionsUser *sql.Stmt
}

// NewHandler creates a new Handler using the provided config.
func NewHandler(ctx context.Context, db *sql.DB, cfg HandlerConfig) (*Handler, error) {
	p := &util.Prepare{
		DB:  db,
		Ctx: ctx,
	}

	h := &Handler{
		providers: make(map[string]IdentityProvider),
		db:        db,

		cfg: cfg,

		updateUser: p.P(`
			update users
			set
				name = case when $2 = '' then name else $2 end,
				email = case when $3 = '' then email else $3 end
			where id = $1
		`),

		userLookup: p.P(`
			select user_id
			from auth_subjects
			where
				provider_id = $1 and
				subject_id = $2
		`),
		addSubject: p.P(`
			insert into auth_subjects (provider_id, subject_id, user_id)
			values ($1, $2, $3)
		`),
		startSession: p.P(`
			insert into auth_user_sessions (id, user_agent, user_id)
			values ($1, $2, $3)
		`),
		endSession: p.P(`
			delete from auth_user_sessions
			where id = any($1)
		`),

		updateUA: p.P(`
			update auth_user_sessions
			set user_agent = $2
			where id = $1
		`),

		fetchSession: p.P(`
			with update as (
				update auth_user_sessions
				set last_access_at = now()
				where id = $1 AND (last_access_at isnull OR last_access_at < now() - '1 minute'::interval)
			)
			select sess.user_id, u.role
			from auth_user_sessions sess
			join users u on u.id = sess.user_id
			where sess.id = $1
		`),

		userSessions: p.P(`
			select id, user_agent, created_at, last_access_at
			from auth_user_sessions
			where user_id = $1
		`),

		endSessionUser: p.P(`
			delete from auth_user_sessions
			where user_id = $1 and id = $2
		`),

		endAllSessionsUser: p.P(`
			delete from auth_user_sessions
			where user_id = $1 and id != $2
		`),
	}

	return h, p.Err
}

// UserSession represents an active user session.
type UserSession struct {
	ID           string
	UserAgent    string
	CreatedAt    time.Time
	LastAccessAt time.Time
	UserID       string
}

func (h *Handler) EndUserSessionTx(ctx context.Context, tx *sql.Tx, id ...string) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}
	if permission.Admin(ctx) {
		_, err = tx.StmtContext(ctx, h.endSession).ExecContext(ctx, sqlutil.UUIDArray(id))
	} else {
		_, err = tx.StmtContext(ctx, h.endSessionUser).ExecContext(ctx, permission.UserNullUUID(ctx), sqlutil.UUIDArray(id))
	}
	return err
}

// EndAllUserSessionsTx ends all sessions other than the user's currently active session
func (h *Handler) EndAllUserSessionsTx(ctx context.Context, tx *sql.Tx) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.MatchUser(permission.UserID(ctx)))
	if err != nil {
		return err
	}

	// get current session id
	src := permission.Source(ctx)

	stmt := h.endAllSessionsUser
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}
	_, err = stmt.ExecContext(ctx, permission.UserNullUUID(ctx), src.ID)

	return err
}

func (h *Handler) FindAllUserSessions(ctx context.Context, userID string) ([]UserSession, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.MatchUser(userID))
	if err != nil {
		return nil, err
	}

	rows, err := h.userSessions.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []UserSession
	for rows.Next() {
		s := UserSession{UserID: userID}
		var lastAccess sql.NullTime
		err = rows.Scan(&s.ID, &s.UserAgent, &s.CreatedAt, &lastAccess)
		if err != nil {
			return nil, err
		}
		s.LastAccessAt = lastAccess.Time.Truncate(time.Minute)
		s.CreatedAt = s.CreatedAt.Truncate(time.Minute)
		sessions = append(sessions, s)
	}

	return sessions, nil
}

// ServeLogout will clear the current session cookie and end the session(s) (if any).
func (h *Handler) ServeLogout(w http.ResponseWriter, req *http.Request) {
	ClearCookie(w, req, CookieName, true)
	var sessionIDs []string
	for _, c := range req.Cookies() {
		switch c.Name {
		case CookieName, v1CookieName:
		default:
			// only interested in cookies with one of the names above
			continue
		}

		tok, _, _ := authtoken.Parse(c.Value, nil)
		if tok == nil {
			continue
		}
		sessionIDs = append(sessionIDs, tok.ID.String())
	}
	ctx := req.Context()
	src := permission.Source(ctx)
	if src != nil && src.Type == permission.SourceTypeAuthProvider {
		sessionIDs = append(sessionIDs, src.ID)
	}

	if len(sessionIDs) == 0 {
		// no session to end
		return
	}

	_, err := h.endSession.ExecContext(log.FromContext(ctx).BackgroundContext(), sqlutil.UUIDArray(sessionIDs))
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "end session(s)"))
	}
}

// ServeProviders will return a list of the currently enabled identity providers.
func (h *Handler) ServeProviders(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := req.Context()
	info := make([]registeredProvider, 0, len(h.providers))

	u, err := url.Parse(req.RequestURI)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
	// Detect current pathPrefix instead of using CallbackURL since it
	// will be used for browser linking.
	//
	// Also handles edge cases around first-time setup/localhost/etc...
	pathPrefix := strings.TrimSuffix(u.Path, req.URL.Path)

	for id, p := range h.providers {
		if !p.Info(ctx).Enabled {
			continue
		}

		info = append(info, registeredProvider{
			ID:           id,
			URL:          path.Join(pathPrefix, "/api/v2/identity/providers", url.PathEscape(id)),
			ProviderInfo: p.Info(ctx),
		})
	}

	sort.Slice(info, func(i, j int) bool { return info[i].ID < info[j].ID })
	data, err := json.Marshal(info)
	if errutil.HTTPError(req.Context(), w, err) {
		return
	}
	_, _ = w.Write(data)
}

// IdentityProviderHandler will return a handler for the given provider ID.
//
// It panics if the id has not been registered with AddIdentityProvider.
func (h *Handler) IdentityProviderHandler(id string) http.HandlerFunc {
	p, ok := h.providers[id]
	if !ok {
		panic("IdentityProvider " + id + " does not exist")
	}

	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		cfg := config.FromContext(ctx)

		var refU *url.URL
		if req.Method == "POST" {
			if cfg.ShouldUsePublicURL() {
				refU, _ = url.Parse(req.Header.Get("referer"))
				if refU == nil || !cfg.ValidReferer("", req.Header.Get("referer")) {
					// redirect with err
					q := make(url.Values)
					q.Set("login_error", "invalid referer")
					http.Redirect(w, req, cfg.CallbackURL("", q), http.StatusTemporaryRedirect)
					return
				}
			} else {
				// fallback to old method
				var ok bool
				refU, ok = h.refererURL(w, req)
				if !ok {
					errutil.HTTPError(ctx, w, validation.NewFieldError("referer", "failed to resolve referer"))
					return
				}
			}
		} else {
			c, err := req.Cookie("login_redir")
			if err != nil {
				errutil.HTTPError(ctx, w, validation.NewFieldError("login_redir", err.Error()))
				return
			}
			refU, _ = url.Parse(c.Value)
			if refU == nil || !cfg.ValidReferer(req.URL.String(), c.Value) {
				// redirect with err
				q := make(url.Values)
				q.Set("login_error", "invalid referer")
				http.Redirect(w, req, cfg.CallbackURL("", q), http.StatusTemporaryRedirect)
				return
			}
		}

		info := p.Info(ctx)
		if !info.Enabled {
			err := Error(info.Title + " auth disabled")
			q := refU.Query()
			q.Set("login_error", err.Error())
			refU.RawQuery = q.Encode()
			http.Redirect(w, req, refU.String(), http.StatusFound)
			return
		}

		if req.Method == "POST" {
			h.serveProviderPost(id, p, refU, w, req)
			return
		}

		h.handleProvider(id, p, refU, w, req)
	}
}

// A Redirector provides a target URL for redirecting a user.
type Redirector interface {
	RedirectURL() string
}

// RedirectURL is a convenience type that can be returned as an error
// resulting in redirection. It implements the error and Redirector interfaces.
type RedirectURL string

// An Error can be returned to indicate an error message that should be displayed to
// the user attempting to authenticate.
type Error string

// ClientError indicates an error meant for the client to see.
func (Error) ClientError() bool { return true }

func (a Error) Error() string { return string(a) }

func (RedirectURL) Error() string { return "must redirect to acquire identity" }

// RedirectURL implements the Redirector interface.
func (r RedirectURL) RedirectURL() string { return string(r) }

func (h *Handler) canCreateUser(ctx context.Context, providerID string) bool {
	cfg := config.FromContext(ctx)
	switch providerID {
	case "oidc":
		return cfg.OIDC.NewUsers
	case "github":
		return cfg.GitHub.NewUsers
	}

	return false
}

func (h *Handler) handleProvider(id string, p IdentityProvider, refU *url.URL, w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var route RouteInfo
	route.RelativePath = strings.TrimPrefix(req.URL.Path, "/v1/identity/providers/"+id)
	route.RelativePath = strings.TrimPrefix(route.RelativePath, "/api/v2/identity/providers/"+id)
	if route.RelativePath == "" {
		route.RelativePath = "/"
	}

	cfg := config.FromContext(ctx)
	if cfg.ShouldUsePublicURL() {
		route.CurrentURL = cfg.CallbackURL(req.URL.Path)
	} else {
		u := *req.URL
		u.RawQuery = "" // strip query params
		route.CurrentURL = u.String()
	}

	sub, err := p.ExtractIdentity(&route, w, req)
	var r Redirector
	if errors.As(err, &r) {
		http.Redirect(w, req, r.RedirectURL(), http.StatusFound)
		return
	}
	noRedirect := req.FormValue("noRedirect") == "1"

	errRedirect := func(err error) {
		q := refU.Query()
		old := err
		_, err = errutil.ScrubError(err)
		if err != old {
			log.Log(ctx, old)
		}
		q.Set("login_error", err.Error())
		refU.RawQuery = q.Encode()
		if noRedirect {
			if err != old {
				w.WriteHeader(500)
			} else {
				w.WriteHeader(400)
			}
			_, _ = io.WriteString(w, err.Error())
			return
		}
		http.Redirect(w, req, refU.String(), http.StatusFound)
	}

	if err != nil {
		errRedirect(err)
		return
	}

	var userID string
	err = h.userLookup.QueryRowContext(ctx, id, sub.SubjectID).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		errRedirect(err)
		return
	}

	var newUser bool
	if userID == "" {
		newUser = true

		if !h.canCreateUser(ctx, id) {
			errRedirect(Error("New users are not allowed right now, but you can try again later."))
			log.Log(ctx, errors.New("create user: disabled for provider"))
			return
		}

		// create user
		tx, err := h.db.BeginTx(ctx, nil)
		if err != nil {
			errRedirect(err)
			return
		}
		defer sqlutil.Rollback(ctx, "auth: create user", tx)

		u := &user.User{
			Role:  permission.RoleUser,
			Name:  validate.SanitizeName(sub.Name),
			Email: validate.SanitizeEmail(sub.Email),
		}
		permission.SudoContext(ctx, func(ctx context.Context) {
			u, err = h.cfg.UserStore.InsertTx(ctx, tx, u)
		})
		if err != nil {
			errRedirect(err)
			return
		}
		_, err = tx.Stmt(h.addSubject).ExecContext(ctx, id, sub.SubjectID, u.ID)
		userID = u.ID
		if err != nil {
			errRedirect(err)
			return
		}
		err = tx.Commit()
		if err != nil {
			errRedirect(err)
			return
		}
	} else {
		_, err = h.updateUser.ExecContext(ctx, userID, validate.SanitizeName(sub.Name),
			validate.SanitizeEmail(sub.Email))
		if err != nil {
			log.Log(ctx, errors.Wrap(err, "update user info"))
		}
	}

	tok, err := h.CreateSession(ctx, req.UserAgent(), userID)
	if err != nil {
		errRedirect(err)
		return
	}
	tokStr, err := tok.Encode(h.cfg.SessionKeyring.Sign)
	if err != nil {
		errRedirect(err)
		return
	}

	if noRedirect {
		_, _ = io.WriteString(w, tokStr)
		return
	}

	h.setSessionCookie(w, req, tokStr)

	if newUser {
		q := refU.Query()
		q.Set("isFirstLogin", "1")
		refU.RawQuery = q.Encode()
	}

	http.Redirect(w, req, refU.String(), http.StatusFound)
}

// CreateSession will start a new session for the given UserID, returning a newly signed token.
func (h *Handler) CreateSession(ctx context.Context, userAgent, userID string) (*authtoken.Token, error) {
	tok := &authtoken.Token{
		Version: 1,
		Type:    authtoken.TypeSession,
		ID:      uuid.New(),
	}
	_, err := h.startSession.ExecContext(ctx, tok.ID.String(), userAgent, userID)
	if err != nil {
		return nil, err
	}

	return tok, nil
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, req *http.Request, val string) {
	SetCookieAge(w, req, CookieName, val, 30*24*time.Hour, true)
}

func (h *Handler) authWithToken(w http.ResponseWriter, req *http.Request, next http.Handler) bool {
	err := req.ParseMultipartForm(32 << 20) // 32<<20 (32MiB) value is the `defaultMaxMemory` used in the net/http package when `req.FormValue` is called
	if err != nil && !errors.Is(err, http.ErrNotMultipart) {
		http.Error(w, err.Error(), 400)
		return true
	}

	tokStr := GetToken(req)
	if tokStr == "" {
		return false
	}

	ctx := req.Context()
	if req.URL.Path == "/api/graphql" && strings.HasPrefix(tokStr, "ey") {
		ctx, err = h.cfg.APIKeyStore.AuthorizeGraphQL(ctx, tokStr, req.UserAgent(), req.RemoteAddr)
		if errutil.HTTPError(req.Context(), w, err) {
			return true
		}

		next.ServeHTTP(w, req.WithContext(ctx))
		return true
	}
	if req.URL.Path == "/api/v2/uik" && strings.HasPrefix(tokStr, "ey") {
		ctx, err = h.cfg.IntKeyStore.AuthorizeUIK(ctx, tokStr)
		if errutil.HTTPError(req.Context(), w, err) {
			return true
		}

		next.ServeHTTP(w, req.WithContext(ctx))
		return true
	}

	tok, _, err := authtoken.Parse(tokStr, func(t authtoken.Type, p, sig []byte) (bool, bool) {
		if t == authtoken.TypeSession {
			return h.cfg.SessionKeyring.Verify(p, sig)
		}

		return h.cfg.APIKeyring.Verify(p, sig)
	})
	if errutil.HTTPError(req.Context(), w, err) {
		return true
	}

	switch req.URL.Path {
	case "/v1/api/alerts", "/api/v2/generic/incoming":
		ctx, err = h.cfg.IntKeyStore.Authorize(ctx, *tok, integrationkey.TypeGeneric)
	case "/v1/webhooks/grafana", "/api/v2/grafana/incoming":
		ctx, err = h.cfg.IntKeyStore.Authorize(ctx, *tok, integrationkey.TypeGrafana)
	case "/api/v2/site24x7/incoming":
		ctx, err = h.cfg.IntKeyStore.Authorize(ctx, *tok, integrationkey.TypeSite24x7)
	case "/api/v2/prometheusalertmanager/incoming":
		ctx, err = h.cfg.IntKeyStore.Authorize(ctx, *tok, integrationkey.TypePrometheusAlertmanager)
	case "/api/v2/calendar":
		ctx, err = h.cfg.CalSubStore.Authorize(ctx, *tok)
	default:
		return false
	}

	if errutil.HTTPError(req.Context(), w, err) {
		return true
	}

	next.ServeHTTP(w, req.WithContext(ctx))
	return true
}

func (h *Handler) tryAuthUser(ctx context.Context, w http.ResponseWriter, req *http.Request, tokenStr string, isCookie bool) (context.Context, error) {
	tok, isOld, err := authtoken.Parse(tokenStr, func(t authtoken.Type, p, sig []byte) (bool, bool) {
		// only session tokens are supported for cookies
		return h.cfg.SessionKeyring.Verify(p, sig)
	})
	if err != nil {
		return nil, err
	}

	var userID uuid.UUID
	var userRole permission.Role
	err = h.fetchSession.QueryRowContext(ctx, tok.ID.String()).Scan(&userID, &userRole)
	if err != nil {
		return nil, err
	}

	if isCookie && isOld {
		// send new signature back if it was signed with an old key
		newSignedToken, err := tok.Encode(h.cfg.SessionKeyring.Sign)
		if err != nil {
			log.Log(ctx, errors.Wrap(err, "failed to sign/issue new session token"))
		} else {
			h.setSessionCookie(w, req, newSignedToken)
			_, err = h.updateUA.ExecContext(ctx, tok.ID.String(), req.UserAgent())
			if err != nil {
				log.Log(ctx, errors.Wrap(err, "update user agent (session key refresh)"))
			}
		}
	}

	return permission.UserSourceContext(
		ctx,
		userID.String(),
		userRole,
		&permission.SourceInfo{
			Type: permission.SourceTypeAuthProvider,
			ID:   tok.ID.String(),
		},
	), nil
}

// WrapHandler will wrap an existing http.Handler so the Context of the request
// includes authentication information (if the request is authorized).
//
// Updating and clearing the session cookie is automatically handled.
func (h *Handler) WrapHandler(wrapped http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/api/v2/slack") {
			wrapped.ServeHTTP(w, req)
			return
		}
		if req.URL.Path == "/api/v2/mailgun/incoming" || req.URL.Path == "/v1/webhooks/mailgun" {
			// Mailgun handles it's own auth and has special
			// requirements on status codes, so we pass it through
			// untouched.
			wrapped.ServeHTTP(w, req)
			return
		}
		if h.authWithToken(w, req, wrapped) {
			return
		}

		// User session flow

		tokStr := GetToken(req)
		// explicit token always takes precedence
		if tokStr != "" {
			ctx, err := h.tryAuthUser(req.Context(), w, req, tokStr, false)
			if err != nil {
				wrapped.ServeHTTP(w, req)
				return
			}

			wrapped.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		for _, c := range req.Cookies() {
			switch c.Name {
			case CookieName, v1CookieName:
			default:
				// only interested in cookies with one of the names above
				continue
			}

			ctx, err := h.tryAuthUser(req.Context(), w, req, c.Value, true)
			if err != nil {
				continue
			}

			wrapped.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		wrapped.ServeHTTP(w, req)
	})
}

func (h *Handler) refererURL(w http.ResponseWriter, req *http.Request) (*url.URL, bool) {
	ref := req.Header.Get("referer")
	ctx := req.Context()
	cfg := config.FromContext(ctx)
	refU, err := url.Parse(ref)
	if err != nil {
		errutil.HTTPError(ctx, w, validation.NewFieldError("referer", err.Error()))
		return nil, false
	}

	if !cfg.ValidReferer(req.URL.String(), ref) {
		err := validation.NewFieldError("referer", "wrong host/path")
		ctx = log.WithFields(ctx, log.Fields{
			"AuthRefererURLs": cfg.Auth.RefererURLs,
			"PublicURL":       cfg.PublicURL(),
		})
		log.Log(ctx, err)
		errutil.HTTPError(ctx, w, err)
		return nil, false
	}

	q := refU.Query() // reset existing login params
	q.Del("isFirstLogin")
	q.Del("login_error")
	refU.RawQuery = q.Encode()
	return refU, true
}

func (h *Handler) serveProviderPost(id string, p IdentityProvider, refU *url.URL, w http.ResponseWriter, req *http.Request) {
	SetCookie(w, req, "login_redir", refU.String(), false)

	h.handleProvider(id, p, refU, w, req)
}

// AddIdentityProvider registers a new IdentityProvider with the given ID.
func (h *Handler) AddIdentityProvider(id string, idp IdentityProvider) error {
	if h.providers[id] != nil {
		return errors.Errorf("provider already exists with id '%s'", id)
	}

	h.providers[id] = idp
	return nil
}
