package auth

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/config"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
	"go.opencensus.io/trace"
)

// CookieName is the name of the auth session cookie.
const CookieName = "goalert_session.2"
const v1CookieName = "goalert_session"

type registeredProvider struct {
	// ID is the unique identifier of the provider.
	ID string

	// URL is the location of the form action (POST).
	URL string

	ProviderInfo
}

// HandlerConfig provides configuration for the auth handler.
type HandlerConfig struct {
	UserStore      user.Store
	SessionKeyring keyring.Keyring
	IntKeyStore    integrationkey.Store
}

// Handler will serve authentication requests for registered identity providers.
type Handler struct {
	providers map[string]IdentityProvider
	cfg       HandlerConfig

	db         *sql.DB
	userLookup *sql.Stmt
	addSubject *sql.Stmt
	updateUA   *sql.Stmt

	startSession *sql.Stmt
	fetchSession *sql.Stmt
	endSession   *sql.Stmt
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
			where id = $1
		`),

		updateUA: p.P(`
			update auth_user_sessions
			set user_agent = $2
			where id = $1
		`),

		fetchSession: p.P(`
			select sess.user_id, u.role
			from auth_user_sessions sess
			join users u on u.id = sess.user_id
			where sess.id = $1
		`),
	}

	return h, p.Err
}

// ServeLogout will clear the current session cookie and end the session (if any).
func (h *Handler) ServeLogout(w http.ResponseWriter, req *http.Request) {
	h.setSessionCookie(w, req, "")
	ctx := req.Context()
	src := permission.Source(ctx)
	if src != nil && src.Type == permission.SourceTypeAuthProvider {
		_, err := h.endSession.ExecContext(context.Background(), src.ID)
		if err != nil {
			log.Log(ctx, errors.Wrap(err, "end session"))
		}
	}
}

// ServeProviders will return a list of the currently enabled identity providers.
func (h *Handler) ServeProviders(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := req.Context()
	info := make([]registeredProvider, 0, len(h.providers))

	for id, p := range h.providers {
		if !p.Info(ctx).Enabled {
			continue
		}

		info = append(info, registeredProvider{
			ID:           id,
			URL:          "/api/v2/identity/providers/" + url.PathEscape(id),
			ProviderInfo: p.Info(ctx),
		})
	}

	sort.Slice(info, func(i, j int) bool { return info[i].ID < info[j].ID })
	data, err := json.Marshal(info)
	if errutil.HTTPError(req.Context(), w, err) {
		return
	}
	w.Write(data)
}

// IdentityProviderHandler will return a handler for the given provider ID.
//
// It panics if the id has not been registerd with AddIdentityProvider.
func (h *Handler) IdentityProviderHandler(id string) http.HandlerFunc {
	p, ok := h.providers[id]
	if !ok {
		panic("IdentityProvider " + id + " does not exist")
	}

	return func(w http.ResponseWriter, req *http.Request) {
		ctx, sp := trace.StartSpan(req.Context(), "Auth.Provider/"+id)
		defer sp.End()

		req = req.WithContext(ctx)

		var refU *url.URL
		if req.Method == "POST" {
			var ok bool
			refU, ok = h.refererURL(w, req)
			if !ok {
				errutil.HTTPError(ctx, w, validation.NewFieldError("referer", "failed to resolve referer"))
				return
			}
		} else {
			c, err := req.Cookie("login_redir")
			if err != nil {
				errutil.HTTPError(ctx, w, validation.NewFieldError("login_redir", err.Error()))
				return
			}
			refU, err = url.Parse(c.Value)
			if err != nil {
				errutil.HTTPError(ctx, w, validation.NewFieldError("login_redir", err.Error()))
				return
			}
		}

		info := p.Info(ctx)
		if !info.Enabled {
			err := Error(info.Title + " auth disabled")
			q := refU.Query()
			sp.Annotate([]trace.Attribute{trace.BoolAttribute("error", true)}, "error: "+err.Error())
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
	sp := trace.FromContext(ctx)

	var route RouteInfo
	route.RelativePath = strings.TrimPrefix(req.URL.Path, "/v1/identity/providers/"+id)
	route.RelativePath = strings.TrimPrefix(route.RelativePath, "/api/v2/identity/providers/"+id)
	if route.RelativePath == "" {
		route.RelativePath = "/"
	}

	u := *req.URL
	u.RawQuery = "" // strip query params
	route.CurrentURL = u.String()

	sub, err := p.ExtractIdentity(&route, w, req)
	if r, ok := err.(Redirector); ok {
		sp.Annotate([]trace.Attribute{trace.StringAttribute("auth.redirectURL", r.RedirectURL())}, "Redirected.")
		http.Redirect(w, req, r.RedirectURL(), http.StatusFound)
		return
	}
	noRedirect := req.FormValue("noRedirect") == "1"

	errRedirect := func(err error) {
		q := refU.Query()
		sp.Annotate([]trace.Attribute{trace.BoolAttribute("error", true)}, "error: "+err.Error())
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
			io.WriteString(w, err.Error())
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
	if err == sql.ErrNoRows {
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
		defer tx.Rollback()
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
		sp.Annotate([]trace.Attribute{
			trace.BoolAttribute("user.new", true),
			trace.StringAttribute("user.name", u.Name),
			trace.StringAttribute("user.email", u.Email),
			trace.StringAttribute("user.id", u.ID),
		}, "Created new user.")
	}

	sessToken, sessID, err := h.CreateSession(ctx, req.UserAgent(), userID)
	if err != nil {
		errRedirect(err)
		return
	}

	sp.Annotate([]trace.Attribute{
		trace.BoolAttribute("auth.login", true),
		trace.StringAttribute("auth.userID", userID),
		trace.StringAttribute("auth.sessionID", sessID),
	}, "User authenticated.")

	if noRedirect {
		io.WriteString(w, sessToken)
		return
	}

	h.setSessionCookie(w, req, sessToken)

	if newUser {
		q := refU.Query()
		q.Set("isFirstLogin", "1")
		refU.RawQuery = q.Encode()
	}

	http.Redirect(w, req, refU.String(), http.StatusFound)
}

// CreateSession will start a new session for the given UserID, returning a newly signed token.
func (h *Handler) CreateSession(ctx context.Context, userAgent, userID string) (token, id string, err error) {
	sessID := uuid.NewV4()
	_, err = h.startSession.ExecContext(ctx, sessID.String(), userAgent, userID)
	if err != nil {
		return "", "", err
	}

	var buf bytes.Buffer
	buf.WriteByte('S') // session IDs will be prefixed with an "S"
	buf.Write(sessID.Bytes())
	sig, err := h.cfg.SessionKeyring.Sign(buf.Bytes())
	if err != nil {
		return "", "", err
	}
	buf.Write(sig)

	return base64.URLEncoding.EncodeToString(buf.Bytes()), sessID.String(), nil
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, req *http.Request, val string) {
	ClearCookie(w, req, "login_redir")
	if val == "" {
		ClearCookie(w, req, CookieName)
	} else {
		SetCookieAge(w, req, CookieName, val, 30*24*time.Hour)
	}
}

func (h *Handler) authWithToken(w http.ResponseWriter, req *http.Request, next http.Handler) bool {
	err := req.ParseMultipartForm(32 << 20) // 32<<20 (32MiB) value is the `defaultMaxMemory` used in the net/http package when `req.FormValue` is called
	if err != nil && err != http.ErrNotMultipart {
		http.Error(w, err.Error(), 400)
		return true
	}

	tok := GetToken(req)
	if tok == "" {
		return false
	}

	// TODO: update once scopes are implemented
	ctx := req.Context()
	switch req.URL.Path {
	case "/v1/api/alerts", "/api/v2/generic/incoming":
		ctx, err = h.cfg.IntKeyStore.Authorize(ctx, tok, integrationkey.TypeGeneric)
	case "/v1/webhooks/grafana", "/api/v2/grafana/incoming":
		ctx, err = h.cfg.IntKeyStore.Authorize(ctx, tok, integrationkey.TypeGrafana)
	case "/v1/webhooks/site24x7", "/api/v2/site24x7/incoming":
		ctx, err = h.cfg.IntKeyStore.Authorize(ctx, tok, integrationkey.TypeSite24x7)
	default:
		return false
	}

	if errutil.HTTPError(req.Context(), w, err) {
		return true
	}

	next.ServeHTTP(w, req.WithContext(ctx))
	return true
}

// WrapHandler will wrap an existing http.Handler so the Context of the request
// includes authentication information (if the request is authorized).
//
// Updating and clearing the session cookie is automatically handled.
func (h *Handler) WrapHandler(wrapped http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
		ctx := req.Context()
		tok := GetToken(req)
		var fromCookie bool
		if tok == "" {
			c, err := req.Cookie(CookieName)
			if err == nil {
				fromCookie = true
				tok = c.Value
			}
		}
		if tok == "" {
			c, err := req.Cookie(v1CookieName)
			if err == nil {
				fromCookie = true
				tok = c.Value
			}
		}

		if tok == "" {
			// no cookie value
			wrapped.ServeHTTP(w, req)
			return
		}
		data, err := base64.URLEncoding.DecodeString(tok)
		if err != nil {
			if fromCookie {
				h.setSessionCookie(w, req, "")
			}
			wrapped.ServeHTTP(w, req)
			return
		}
		if len(data) == 0 || data[0] != 'S' || len(data) < 17 {
			if fromCookie {
				h.setSessionCookie(w, req, "")
			}
			wrapped.ServeHTTP(w, req)
			return
		}

		id, err := uuid.FromBytes(data[1:17])
		if err != nil {
			if fromCookie {
				h.setSessionCookie(w, req, "")
			}
			wrapped.ServeHTTP(w, req)
			return
		}

		valid, isOld := h.cfg.SessionKeyring.Verify(data[:17], data[17:])
		if !valid {
			if fromCookie {
				h.setSessionCookie(w, req, "")
			}
			wrapped.ServeHTTP(w, req)
			return
		}
		if fromCookie && isOld {
			// send new signature back if it was signed with an old key
			sig, err := h.cfg.SessionKeyring.Sign(data[:17])
			if err != nil {
				log.Log(ctx, errors.Wrap(err, "failed to sign/issue new session token"))
			} else {
				data = append(data[:17], sig...)
				h.setSessionCookie(w, req, base64.URLEncoding.EncodeToString(data))

				_, err = h.updateUA.ExecContext(ctx, id.String(), req.UserAgent())
				if err != nil {
					log.Log(ctx, errors.Wrap(err, "update user agent (session key refresh)"))
				}
			}
		} else if fromCookie {
			// compat, always set cookie (for transition from /v1 to /api)
			h.setSessionCookie(w, req, tok)
		}

		var userID string
		var userRole permission.Role
		err = h.fetchSession.QueryRowContext(ctx, id.String()).Scan(&userID, &userRole)
		if err == sql.ErrNoRows {
			if fromCookie {
				h.setSessionCookie(w, req, "")
			}
			wrapped.ServeHTTP(w, req)
			return
		}
		if err != nil {
			errutil.HTTPError(ctx, w, err)
			return
		}

		ctx = permission.UserSourceContext(
			ctx,
			userID,
			userRole,
			&permission.SourceInfo{
				Type: permission.SourceTypeAuthProvider,
				ID:   id.String(),
			},
		)
		req = req.WithContext(ctx)

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
	SetCookie(w, req, "login_redir", refU.String())

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
