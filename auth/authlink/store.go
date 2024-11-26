package authlink

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

type Store struct {
	db *sql.DB

	k keyring.Keyring
}

type Metadata struct {
	UserDetails string
	AlertID     int    `json:",omitempty"`
	AlertAction string `json:",omitempty"`
}

func (m Metadata) Validate() error {
	return validate.Many(
		validate.ASCII("UserDetails", m.UserDetails, 1, 255),
		validate.OneOf("AlertAction", m.AlertAction, "", "ResultAcknowledge", "ResultResolve"),
	)
}

func NewStore(ctx context.Context, db *sql.DB, k keyring.Keyring) (*Store, error) {
	p := &util.Prepare{
		DB:  db,
		Ctx: ctx,
	}

	return &Store{
		db: db,
		k:  k,
	}, p.Err
}

func (s *Store) FindLinkMetadata(ctx context.Context, token string) (*Metadata, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	tokID, err := s.tokenID(ctx, token)
	if err != nil {
		// don't return anything, treat it as not found
		return nil, nil
	}

	data, err := gadb.NewCompat(s.db).AuthLinkMetadata(ctx, uuid.MustParse(tokID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var meta Metadata
	err = json.Unmarshal(data, &meta)
	if err != nil {
		return nil, err
	}

	return &meta, nil
}

func (s *Store) tokenID(ctx context.Context, token string) (string, error) {
	var c jwt.RegisteredClaims
	_, err := s.k.VerifyJWT(token, &c, "goalert", "auth-link")
	if err != nil {
		return "", validation.WrapError(err)
	}

	err = validate.UUID("ID", c.ID)
	if err != nil {
		return "", err
	}

	return c.ID, nil
}

func (s *Store) LinkAccount(ctx context.Context, token string) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	tokID, err := s.tokenID(ctx, token)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer sqlutil.Rollback(ctx, "authlink: link auth subject", tx)

	row, err := gadb.NewCompat(tx).AuthLinkUseReq(ctx, uuid.MustParse(tokID))
	if errors.Is(err, sql.ErrNoRows) {
		return validation.NewGenericError("invalid link token")
	}
	if err != nil {
		return err
	}

	err = gadb.NewCompat(tx).AuthLinkAddAuthSubject(ctx, gadb.AuthLinkAddAuthSubjectParams{
		ProviderID: row.ProviderID,
		SubjectID:  row.SubjectID,
		UserID:     uuid.MustParse(permission.UserID(ctx)),
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) AuthLinkURL(ctx context.Context, providerID, subjectID string, meta Metadata) (string, error) {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return "", err
	}
	err = validate.Many(
		validate.SubjectID("ProviderID", providerID),
		validate.SubjectID("SubjectID", subjectID),
		meta.Validate(),
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
	c.NotBefore = jwt.NewNumericDate(now.Add(-2 * time.Minute))
	c.ExpiresAt = jwt.NewNumericDate(expires)
	c.IssuedAt = jwt.NewNumericDate(now)

	token, err := s.k.SignJWT(c)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(meta)
	if err != nil {
		return "", err
	}
	err = gadb.NewCompat(s.db).AuthLinkAddReq(ctx, gadb.AuthLinkAddReqParams{
		ID:         id,
		ProviderID: providerID,
		SubjectID:  subjectID,
		ExpiresAt:  pgtype.Timestamptz{Time: expires, Valid: true},
		Metadata:   data,
	})
	if err != nil {
		return "", err
	}

	cfg := config.FromContext(ctx)
	p := make(url.Values)
	p.Set("authLinkToken", token)
	return cfg.CallbackURL("/profile", p), nil
}
