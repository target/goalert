package authlink

import (
	"context"
	"crypto/rand"
	"database/sql"
	"math/big"
	"sync"
	"time"

	"github.com/jackc/pgtype"
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
)

const (
	MaxClaimQueryFrequency = 3 * time.Second
	ExpirationDuration     = 10 * time.Minute
)

type Store struct {
	db *sql.DB

	unclaimed *sql.Stmt

	create *sql.Stmt
	claim  *sql.Stmt
	verify *sql.Stmt
	auth   *sql.Stmt

	reset  *sql.Stmt
	status *sql.Stmt

	mx           sync.Mutex
	closeCh      chan struct{}
	waitRefresh  chan struct{}
	startRefresh chan struct{}
	wl           *Whitelist

	keys keyring.Keyring
}

type Status struct {
	ID string

	ClaimCode  string
	VerifyCode string

	CreatedAt  time.Time
	ExpiresAt  time.Time
	ClaimedAt  time.Time
	VerifiedAt time.Time
	AuthedAt   time.Time
}

const codeAlphabet = "ABCDEFGHJKLMNPQRSTWXYZ23456789"

func genCode(n int) (string, error) {
	val := make([]byte, n)
	for i := range val {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(codeAlphabet))))
		if err != nil {
			return "", err
		}
		val[i] = codeAlphabet[num.Int64()]
	}

	return string(val), nil
}

func NewStore(ctx context.Context, db *sql.DB, keys keyring.Keyring) (*Store, error) {
	p := util.Prepare{DB: db, Ctx: ctx}
	s := &Store{
		wl:           NewWhitelist(),
		keys:         keys,
		closeCh:      make(chan struct{}),
		waitRefresh:  make(chan struct{}),
		startRefresh: make(chan struct{}),

		unclaimed: p.P(`
			select claim_code
			from auth_link_codes
			where
				now() < expires_at and
				claimed_at isnull and
				verified_at isnull and
				authed_at isnull
		`),

		create: p.P(`
			insert into auth_link_codes (id, user_id, auth_user_session_id, claim_code, verify_code, expires_at)
			values ($1, $2, $3, $4, $5, now() + $6::interval)
			returning created_at
		`),
		claim: p.P(`
			update auth_link_codes set claimed_at = now()
			where
				claim_code = $1 and
				now() < expires_at and
				claimed_at isnull and
				verified_at isnull and
				authed_at isnull
			returning id, claimed_at, verify_code, expires_at
		`),
		verify: p.P(`
			update auth_link_codes
			set verified_at = now()
			where
				id = $1 and
				user_id = $2 and
				auth_user_session_id = $3
				and verify_code = $4 and
				now() < expires_at and
				claimed_at notnull and
				verified_at isnull and
				authed_at isnull
		`),
		auth: p.P(`
			update auth_link_codes
			set authed_at = now()
			where
				id = $1 and
				now() < expires_at and
				claimed_at notnull and
				verified_at notnull and
				authed_at isnull
			returning user_id
		`),

		reset: p.P(`delete from auth_link_codes where user_id = $1`),
		status: p.P(`
			select created_at, expires_at, claimed_at, verified_at, authed_at
			from auth_link_codes
			where id = $1 and user_id = $2 and auth_user_session_id = $3
		`),
	}
	if p.Err != nil {
		return nil, p.Err
	}
	go s.claimUpdates()

	return s, nil
}

func withTx(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt) *sql.Stmt {
	if tx == nil {
		return stmt
	}

	return tx.StmtContext(ctx, stmt)
}

// CreateTx will create a new auth link for the session associated with ctx.
func (s *Store) CreateTx(ctx context.Context, tx *sql.Tx) (*Status, error) {
	err := permission.LimitCheckAny(ctx, permission.UserSession)
	if err != nil {
		return nil, err
	}
	stat := Status{ID: uuid.NewV4().String()}
	stat.ClaimCode, err = genCode(8)
	if err != nil {
		return nil, err
	}
	stat.VerifyCode, err = genCode(8)
	if err != nil {
		return nil, err
	}
	var dur pgtype.Interval
	dur.Set(ExpirationDuration)
	err = withTx(ctx, tx, s.create).QueryRowContext(ctx,
		stat.ID,
		permission.UserID(ctx),
		permission.SessionID(ctx),
		stat.ClaimCode,
		stat.VerifyCode,
		dur,
	).Scan(&stat.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &stat, nil
}

// StatusByID will return the current auth link status for the given ID.
func (s *Store) StatusByID(ctx context.Context, id string) (*Status, error) {
	err := permission.LimitCheckAny(ctx, permission.UserSession)
	if err != nil {
		return nil, err
	}

	var claim, verify, auth sql.NullTime
	stat := &Status{ID: id}
	err = s.status.QueryRowContext(ctx, id, permission.UserID(ctx), permission.SessionID(ctx)).Scan(
		&stat.CreatedAt,
		&stat.ExpiresAt,
		&claim, &verify, &auth,
	)
	if err != nil {
		return nil, err
	}

	stat.ClaimedAt = claim.Time
	stat.VerifiedAt = verify.Time
	stat.AuthedAt = auth.Time

	return stat, nil
}

// Verify will perform verification on the provided auth link with the given code.
func (s *Store) Verify(ctx context.Context, id, code string) (bool, error) {
	err := permission.LimitCheckAny(ctx, permission.UserSession)
	if err != nil {
		return false, err
	}

	rows, err := s.verify.ExecContext(ctx, id, permission.UserID(ctx), permission.SessionID(ctx), code)
	if err != nil {
		return false, err
	}
	n, err := rows.RowsAffected()
	if err != nil {
		return false, err
	}

	return n != 0, nil
}

// ResetTx will reset any outstanding auth links for the user associated with ctx.
func (s *Store) ResetTx(ctx context.Context, tx *sql.Tx) error {
	err := permission.LimitCheckAny(ctx, permission.UserSession)
	if err != nil {
		return nil
	}

	_, err = withTx(ctx, tx, s.reset).ExecContext(ctx, permission.UserID(ctx))
	return err
}
