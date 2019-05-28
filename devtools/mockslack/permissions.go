package mockslack

import (
	"context"
	"net/http"
	"strings"
)

type contextKey int

const (
	contextKeyToken contextKey = iota
)

// AuthToken represents a state of authorization with the Slack server.
type AuthToken struct {
	ID     string
	Scopes []string
	User   string
	IsBot  bool
}

// WithToken will return a new context authorized for API calls with the given AuthToken.
func WithToken(ctx context.Context, tok *AuthToken) context.Context {
	if tok == nil {
		return ctx
	}
	cpy := *tok
	cpy.Scopes = make([]string, len(tok.Scopes))
	copy(cpy.Scopes, tok.Scopes)
	return context.WithValue(ctx, contextKeyToken, cpy)
}

// ContextToken will return a copy of the AuthToken from the given context.
func ContextToken(ctx context.Context) *AuthToken {
	tok, ok := ctx.Value(contextKeyToken).(AuthToken)
	if !ok {
		return nil
	}
	return &tok
}

func tokenID(ctx context.Context) string {
	tok := ContextToken(ctx)
	if tok == nil {
		return ""
	}
	return tok.ID
}
func userID(ctx context.Context) string {
	tok := ContextToken(ctx)
	if tok == nil || tok.IsBot {
		return ""
	}

	return tok.User
}
func botID(ctx context.Context) string {
	tok := ContextToken(ctx)
	if tok == nil || !tok.IsBot {
		return ""
	}

	return tok.User
}

type scopeError struct {
	response
	Needed   string `json:"needed"`
	Provided string `json:"provided"`
}

func (r response) Error() string { return r.Err }

func hasScope(ctx context.Context, scopes ...string) bool {
	return ContextToken(ctx).hasScope(scopes...)
}

func (tok *AuthToken) hasScope(scopes ...string) bool {
	if tok == nil {
		return false
	}
	for _, scope := range scopes {
		for _, tokenScope := range tok.Scopes {
			if scope == tokenScope {
				return true
			}
		}
	}

	return false
}

func checkPermission(ctx context.Context, scopes ...string) error {
	tok := ContextToken(ctx)
	if tok == nil {
		return response{Err: "not_authed"}
	}

	if len(scopes) == 0 || tok.hasScope(scopes...) {
		return nil
	}

	return &scopeError{
		Needed:   strings.Join(scopes, ","),
		Provided: strings.Join(tok.Scopes, ","),
		response: response{Err: "missing_scope"},
	}
}

func (s *Server) tokenMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		tok := s.token(req.FormValue("token"))
		if tok == nil && strings.HasPrefix(req.Header.Get("Authorization"), "Bearer ") {
			tok = s.token(strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer "))
		}
		if c, _ := req.Cookie(TokenCookieName); tok == nil && c != nil {
			tok = s.token(c.Value)
		}

		next(w, req.WithContext(WithToken(req.Context(), tok)))
	}
}
