package integrationkey

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/target/goalert/auth/authtoken"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

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

	tok := authtoken.Token{
		Version: 3,
		Type:    authtoken.TypeUIK,
		ID:      uuid.New(),
	}
	tokStr, err := tok.Encode(s.keys.Sign)
	if err != nil {
		return "", err
	}
	hint := tokStr[:3] + "..." + tokStr[len(tokStr)-4:]

	err = s.setToken(ctx, db, id, tok.ID, hint)
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
