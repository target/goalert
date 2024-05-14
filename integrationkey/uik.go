package integrationkey

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

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

	// TODO: fetch rules & config & process
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
		Type: permission.SourceTypeIntegrationKey,
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
