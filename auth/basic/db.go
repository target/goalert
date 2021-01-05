package basic

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// Store can create new user/pass links and validate a username and password. bcrypt is used
// for password storage & verification.
type Store struct {
	insert        *sql.Stmt
	getByUsername *sql.Stmt
}

const tableName = "auth_basic_users"
const passCost = 14

// NewStore creates a new DB. Error is returned if the prepared statements fail to register.
func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{
		DB:  db,
		Ctx: ctx,
	}
	return &Store{
		insert:        p.P(fmt.Sprintf("INSERT INTO %s(user_id, username, password_hash) VALUES ($1, $2, $3)", tableName)),
		getByUsername: p.P(fmt.Sprintf("SELECT user_id, password_hash FROM %s WHERE username = $1", tableName)),
	}, p.Err
}

// CreateTx should add a new entry for the username/password combination linking to userID.
// An error is returned if the username is not unique or the userID is invalid.
// Must have same user or admin role.
func (b *Store) CreateTx(ctx context.Context, tx *sql.Tx, userID, username, password string) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.MatchUser(userID))
	if err != nil {
		return err
	}

	err = validate.Many(
		validate.UUID("UserID", userID),
		validate.UserName("UserName", username),
		validate.Text("Password", password, 8, 200),
	)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), passCost)
	if err != nil {
		return err
	}
	_, err = tx.StmtContext(ctx, b.insert).ExecContext(ctx, userID, username, string(hashedPassword))
	return err
}

// Validate should return a userID if the username and password match.
func (b *Store) Validate(ctx context.Context, username, password string) (string, error) {
	err := validate.Many(
		validate.UserName("UserName", username),
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

	err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	if err != nil {
		return "", errors.WithMessage(err, "invalid password")
	}

	return userID, nil
}
