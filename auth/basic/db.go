package basic

import (
	"context"
	"database/sql"
	"sync"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// Store can create new user/pass links and validate a username and password. bcrypt is used
// for password storage & verification.
type Store struct {
	insert        *sql.Stmt
	getByUsername *sql.Stmt
	getByID       *sql.Stmt
	update        *sql.Stmt

	mx sync.Mutex
}

// NewStore creates a new DB. Error is returned if the prepared statements fail to register.
func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{
		DB:  db,
		Ctx: ctx,
	}
	return &Store{
		insert:        p.P("INSERT INTO auth_basic_users (user_id, username, password_hash) VALUES ($1, $2, $3)"),
		getByUsername: p.P("SELECT user_id, password_hash FROM auth_basic_users WHERE username = $1"),
		getByID:       p.P("SELECT password_hash FROM auth_basic_users WHERE user_id = $1"),
		update:        p.P("UPDATE auth_basic_users SET password_hash = $2 WHERE user_id = $1"),
	}, p.Err
}

// HashedPassword is an interface that can be used to store a password.
type HashedPassword interface {
	Hash() string

	_private() // prevent external implementations
}

type ValidatedPassword interface {
	Hash() string

	_private() // prevent external implementations
}

type hashed []byte

func (h hashed) Hash() string { return string(h) }
func (h hashed) _private()    {}

// NewHashedPassword will hash the given password and return a Password object.
func (b *Store) NewHashedPassword(password string) (HashedPassword, error) {
	err := validate.Text("Password", password, 8, 200)
	if err != nil {
		return nil, err
	}

	b.mx.Lock()
	defer b.mx.Unlock()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	return hashed(hashedPassword), nil
}

// CreateTx should add a new entry for the username/password combination linking to userID.
// An error is returned if the username is not unique or the userID is invalid.
// Must have same user or admin role.
func (b *Store) CreateTx(ctx context.Context, tx *sql.Tx, userID, username string, password HashedPassword) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.MatchUser(userID))
	if err != nil {
		return err
	}

	err = validate.Many(
		validate.UUID("UserID", userID),
		validate.Username("Username", username),
	)
	if err != nil {
		return err
	}

	_, err = tx.StmtContext(ctx, b.insert).ExecContext(ctx, userID, username, password.Hash())
	return err
}

// UpdateTx should update a user's password in auth_basic_users
func (b *Store) UpdateTx(ctx context.Context, tx *sql.Tx, userID string, oldPassword ValidatedPassword, newPassword HashedPassword) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.MatchUser(userID))
	if err != nil {
		return err
	}

	if oldPassword == nil && !permission.Admin(ctx) {
		return validation.NewFieldError("oldPassword", "Previous password required")
	}

	_, err = tx.StmtContext(ctx, b.update).ExecContext(ctx, userID, newPassword.Hash())
	return err
}

// Validate should return a userID if the username and password match.
func (b *Store) Validate(ctx context.Context, username, password string) (string, error) {
	err := validate.Many(
		validate.Username("Username", username),
		validate.Text("Password", password, 1, 200),
	)
	if err != nil {
		return "", err
	}

	row := b.getByUsername.QueryRowContext(ctx, username)
	var userID, hashed string
	err = row.Scan(&userID, &hashed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("invalid username")
		}
		return "", errors.WithMessage(err, "user lookup failure")
	}

	// Since this can be CPU intensive, we'll only allow one at a time.
	b.mx.Lock()
	defer b.mx.Unlock()

	err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	if err != nil {
		return "", errors.WithMessage(err, "invalid password")
	}

	return userID, nil
}

// ValidatePassword should check if the password matches the user's stored password
func (b *Store) ValidatePassword(ctx context.Context, userID string, password string) (ValidatedPassword, error) {
	err := validate.Many(
		validate.UUID("UserID", userID),
		validate.Text("oldPassword", password, 8, 200),
	)
	if err != nil {
		return nil, err
	}

	row := b.getByID.QueryRowContext(ctx, userID)
	var hash string
	err = row.Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid userID")
		}
		return nil, errors.WithMessage(err, "user lookup failure")
	}

	b.mx.Lock()
	defer b.mx.Unlock()

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return nil, validation.NewFieldError("oldPassword", "invalid password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	return hashed(hashedPassword), nil
}
