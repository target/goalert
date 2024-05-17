package integrationkey

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/expr-lang/expr"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"

	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// Issuer is the JWT issuer for UIK API keys.
const Issuer = "goalert"

// Audience is the JWT audience for UIK API keys.
const Audience = "uik-key-v1"

func newClaims(keyID, tokenID uuid.UUID) jwt.Claims {
	n := time.Now()
	return jwt.RegisteredClaims{
		ID:        tokenID.String(),
		Subject:   keyID.String(),
		IssuedAt:  jwt.NewNumericDate(n),
		NotBefore: jwt.NewNumericDate(n.Add(-time.Minute)),
		Issuer:    Issuer,
		Audience:  []string{Audience},
	}
}

func (s *Store) HandleUIK(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if !expflag.ContextHas(ctx, expflag.UnivKeys) {
		errutil.HTTPError(ctx, w, validation.NewGenericError("universal keys are disabled"))
		return
	}

	err := permission.LimitCheckAny(req.Context(), permission.Service)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
	src := permission.Source(ctx)
	if src.Type != permission.SourceTypeUIK {
		// we don't want to allow regular API keys to be used here
		errutil.HTTPError(ctx, w, permission.Unauthorized())
		return
	}

	keyID, err := uuid.Parse(src.ID)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
	data, err := io.ReadAll(req.Body)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
	var body any
	err = json.Unmarshal(data, &body)
	if errutil.HTTPError(ctx, w, validation.WrapError(err)) {
		return
	}

	cfg, err := s.Config(ctx, s.db, keyID)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	env := map[string]any{
		"sprintf": fmt.Sprintf,
		"req": map[string]any{
			"body": body,
		},
	}

	// We need to track if any rule matched, so we can apply default actions if none did.
	var anyMatched bool
	var results []ActionResult
	for _, rule := range cfg.Rules {
		result, err := expr.Eval(rule.ConditionExpr, env)
		if errutil.HTTPError(ctx, w, validation.WrapError(err)) {
			return
		}
		r, ok := result.(bool)
		if !ok {
			errutil.HTTPError(ctx, w, validation.NewGenericError("condition expression must return a boolean"))
			return
		}
		anyMatched = anyMatched || r
		if !r {
			continue
		}

		for _, action := range rule.Actions {
			res := ActionResult{
				DestType: action.Type,
				Values:   action.StaticParams,
				Params:   make(map[string]string, len(action.DynamicParams)),
			}

			for name, exprStr := range action.DynamicParams {
				val, err := expr.Eval(exprStr, env)
				if errutil.HTTPError(ctx, w, validation.WrapError(err)) {
					return
				}
				if _, ok := val.(string); !ok {
					errutil.HTTPError(ctx, w, validation.NewGenericError("dynamic param expressions must return a string"))
					return
				}
				res.Params[name] = val.(string)
			}
			results = append(results, res)
		}
	}

	if !anyMatched {
		// Default actions need to be applied if no rules matched (or if there are no rules at all).
		for _, action := range cfg.DefaultActions {
			res := ActionResult{
				DestType: action.Type,
				Values:   action.StaticParams,
				Params:   make(map[string]string, len(action.DynamicParams)),
			}

			for name, exprStr := range action.DynamicParams {
				val, err := expr.Eval(exprStr, env)
				if errutil.HTTPError(ctx, w, validation.WrapError(err)) {
					return
				}
				if _, ok := val.(string); !ok {
					errutil.HTTPError(ctx, w, validation.NewGenericError("dynamic param expressions must return a string"))
					return
				}
				res.Params[name] = val.(string)
			}
			results = append(results, res)
		}
	}

	log.Logf(ctx, "uik: action result: %#v", results)

	w.WriteHeader(http.StatusNoContent)
}

type ActionResult struct {
	DestType string
	Values   map[string]string
	Params   map[string]string
}

func (s *Store) AuthorizeUIK(ctx context.Context, tokStr string) (context.Context, error) {
	if !expflag.ContextHas(ctx, expflag.UnivKeys) {
		return ctx, permission.Unauthorized()
	}

	var claims jwt.RegisteredClaims
	_, err := s.keys.VerifyJWT(tokStr, &claims, Issuer, Audience)
	if err != nil {
		return ctx, permission.Unauthorized()
	}

	keyID, err := uuid.Parse(claims.Subject)
	if err != nil {
		log.Logf(ctx, "apikey: invalid subject: %v", err)
		return ctx, permission.Unauthorized()
	}
	tokID, err := uuid.Parse(claims.ID)
	if err != nil {
		log.Logf(ctx, "apikey: invalid token ID: %v", err)
		return ctx, permission.Unauthorized()
	}

	serviceID, err := gadb.New(s.db).IntKeyUIKValidateService(ctx, gadb.IntKeyUIKValidateServiceParams{
		KeyID:   keyID,
		TokenID: uuid.NullUUID{UUID: tokID, Valid: true},
	})
	if errors.Is(err, sql.ErrNoRows) {
		return ctx, permission.Unauthorized()
	}
	if err != nil {
		return ctx, err
	}

	ctx = permission.ServiceSourceContext(ctx, serviceID.String(), &permission.SourceInfo{
		Type: permission.SourceTypeUIK,
		ID:   keyID.String(),
	})

	return ctx, nil
}

func (s *Store) TokenHints(ctx context.Context, db gadb.DBTX, id uuid.UUID) (primary, secondary string, err error) {
	err = permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return "", "", err
	}

	row, err := gadb.New(s.db).IntKeyTokenHints(ctx, id)
	// if errors.Is(err, sql.ErrNoRows) {
	// 	return "", "", err
	// }
	if err != nil {
		return "", "", err
	}

	return row.PrimaryTokenHint.String, row.SecondaryTokenHint.String, nil
}

func (s *Store) GenerateToken(ctx context.Context, db gadb.DBTX, id uuid.UUID) (string, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return "", err
	}

	key, err := gadb.New(db).IntKeyFindOne(ctx, id)
	if err != nil {
		return "", err
	}
	if key.Type != gadb.EnumIntegrationKeysType(TypeUniversal) {
		return "", validation.NewFieldError("ID", "key is not a universal key")
	}

	tokID := uuid.New()
	tokStr, err := s.keys.SignJWT(newClaims(id, tokID))
	if err != nil {
		return "", err
	}

	hint := tokStr[:2] + "..." + tokStr[len(tokStr)-4:]

	err = s.setToken(ctx, db, id, tokID, hint)
	if err != nil {
		return "", err
	}

	return tokStr, nil
}

func (s *Store) setToken(ctx context.Context, db gadb.DBTX, keyID, tokenID uuid.UUID, tokenHint string) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	if err := validate.ASCII("Token Hint", tokenHint, 1, 32); err != nil {
		// not user specified, so return a generic error
		return errors.New("invalid token hint")
	}

	gdb := gadb.New(db)

	_, err = gdb.IntKeySetSecondaryToken(ctx, gadb.IntKeySetSecondaryTokenParams{
		ID:                 keyID,
		SecondaryToken:     uuid.NullUUID{UUID: tokenID, Valid: true},
		SecondaryTokenHint: sql.NullString{String: tokenHint, Valid: true},
	})
	if errors.Is(err, sql.ErrNoRows) {
		// it's possible there was never a primary token set
		_, err = gdb.IntKeySetPrimaryToken(ctx, gadb.IntKeySetPrimaryTokenParams{
			ID:               keyID,
			PrimaryToken:     uuid.NullUUID{UUID: tokenID, Valid: true},
			PrimaryTokenHint: sql.NullString{String: tokenHint, Valid: true},
		})
		if errors.Is(err, sql.ErrNoRows) {
			return validation.NewGenericError("key not found, or already has primary and secondary tokens")
		}

		// Note: A possible race condition here is if multiple requests are made to set the primary token at the same time. In that case, the first request will win and the others will fail with a unique constraint violation. This is acceptable because the primary token is only set once, and only rotated thereafter.
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) PromoteSecondaryToken(ctx context.Context, db gadb.DBTX, id uuid.UUID) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	hint, err := gadb.New(db).IntKeyPromoteSecondary(ctx, id)
	if err != nil {
		return err
	}

	if !hint.Valid {
		return validation.NewGenericError("no secondary token to promote")
	}

	return nil
}

func (s *Store) DeleteSecondaryToken(ctx context.Context, db gadb.DBTX, id uuid.UUID) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	err = gadb.New(db).IntKeyDeleteSecondaryToken(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
