package limit

import (
	"context"
	"database/sql"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"
)

// A Store allows getting and setting system limits.
type Store interface {
	// ResetAll will reset all configurable limits to the default (no-limit).
	ResetAll(context.Context) error

	// Max will return the current max value for the given limit.
	Max(context.Context, ID) (int, error)

	// SetMax allows setting the max value for a limit.
	SetMax(context.Context, ID, int) error

	// All will get the current value of all limits.
	All(context.Context) (Limits, error)
}

// DB implements the Store interface against a Postgres DB.
type DB struct {
	findAll  *sql.Stmt
	findOne  *sql.Stmt
	setOne   *sql.Stmt
	resetAll *sql.Stmt
}

// NewDB creates a new DB and prepares all necessary SQL statements.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}
	return &DB{
		findAll: p.P(`select id, max from config_limits`),
		findOne: p.P(`select max from config_limits where id = $1`),
		setOne: p.P(`
			insert into config_limits (id, max)
			values ($1, $2)
			on conflict (id) do update
			set max = $2
		`),
		resetAll: p.P(`truncate config_limits`),
	}, p.Err
}

// ResetAll implements the Store interface.
func (db *DB) ResetAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return err
	}
	_, err = db.resetAll.ExecContext(ctx)
	return err
}

// Max implements the Store interface.
func (db *DB) Max(ctx context.Context, id ID) (int, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return 0, err
	}
	err = id.Valid()
	if err != nil {
		return 0, err
	}
	var max int
	err = db.findOne.QueryRowContext(ctx, id).Scan(&max)
	if err == sql.ErrNoRows {
		return -1, nil
	}
	if err != nil {
		return 0, err
	}
	return max, nil
}

// All implements the Store interface.
func (db *DB) All(ctx context.Context) (Limits, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}
	rows, err := db.findAll.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var id string
	var max int
	l := make(Limits, 8)
	for rows.Next() {
		err = rows.Scan(&id, max)
		if err != nil {
			return nil, err
		}
		l[ID(id)] = max
	}
	return l, nil
}

// SetMax implements the Store interface.
func (db *DB) SetMax(ctx context.Context, id ID, max int) error {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return err
	}
	err = validate.Many(id.Valid(), validate.Range("Max", max, -1, 9000))
	if err != nil {
		return err
	}

	_, err = db.setOne.ExecContext(ctx, id, max)
	return err
}
