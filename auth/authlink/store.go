package authlink

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/target/goalert/config"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

type Store struct {
	db *sql.DB

	k keyring.Keyring

	newLink    *sql.Stmt
	rmLink     *sql.Stmt
	addSubject *sql.Stmt
}

func NewStore(ctx context.Context, db *sql.DB, k keyring.Keyring) (*Store, error) {
	p := &util.Prepare{
		DB:  db,
		Ctx: ctx,
	}

	return &Store{
		db:         db,
		k:          k,
		newLink:    p.P(`insert into auth_link_requests (id, provider_id, subject_id, expires_at) values ($1, $2, $3, $4)`),
		rmLink:     p.P(`delete from auth_link_requests where id = $1 and expires_at > now() returning provider_id, subject_id`),
		addSubject: p.P(`insert into auth_subjects (provider_id, subject_id, user_id) values ($1, $2, $3)`),
	}, p.Err
}

func (s *Store) LinkAccount(ctx context.Context, token string) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	var c jwt.RegisteredClaims
	_, err = s.k.VerifyJWT(token, &c)
	if err != nil {
		return validation.WrapError(err)
	}

	if !c.VerifyIssuer("goalert", true) {
		return validation.NewGenericError("invalid issuer")
	}
	if !c.VerifyAudience("auth-link", true) {
		return validation.NewGenericError("invalid audience")
	}
	err = validate.UUID("ID", c.ID)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var providerID, subjectID string
	err = tx.StmtContext(ctx, s.rmLink).QueryRowContext(ctx, c.ID).Scan(&providerID, &subjectID)
	if errors.Is(err, sql.ErrNoRows) {
		return validation.NewGenericError("invalid link token")
	}
	if err != nil {
		return err
	}

	_, err = tx.StmtContext(ctx, s.addSubject).ExecContext(ctx, providerID, subjectID, permission.UserID(ctx))
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) AuthLinkURL(ctx context.Context, providerID, subjectID string) (string, error) {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return "", err
	}
	err = validate.Many(
		validate.SubjectID("ProviderID", providerID),
		validate.SubjectID("SubjectID", subjectID),
	)
	if err != nil {
		return "", err
	}

	id := uuid.New()
	now := time.Now()
	expires := now.Add(5 * time.Minute)

	var c jwt.RegisteredClaims
	c.ID = id.String()
	c.Audience = jwt.ClaimStrings{"auth-link"}
	c.Issuer = "goalert"
	c.NotBefore = jwt.NewNumericDate(now.Add(-5 * time.Minute))
	c.ExpiresAt = jwt.NewNumericDate(expires)
	c.IssuedAt = jwt.NewNumericDate(now)

	token, err := s.k.SignJWT(c)
	if err != nil {
		return "", err
	}

	_, err = s.newLink.ExecContext(ctx, id, providerID, subjectID, expires)
	if err != nil {
		return "", err
	}

	cfg := config.FromContext(ctx)
	p := make(url.Values)
	p.Set("authLinkToken", token)
	return cfg.CallbackURL("/profile", p), nil
}
