package limit

import (
	"context"
	"database/sql"
	"errors"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"
)

// A Store allows getting and setting system limits.
type Store struct {
	update   *sql.Stmt
	findAll  *sql.Stmt
	findOne  *sql.Stmt
	setOne   *sql.Stmt
	resetAll *sql.Stmt
}

// NewStore creates a new DB and prepares all necessary SQL statements.
func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}
	return &Store{
		update:  p.P(`update config_limits set max = $2 where id = $1`),
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

// UpdateLimitsTx updates all configurable limits.
func (s *Store) UpdateLimitsTx(ctx context.Context, tx *sql.Tx, id string, max int) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return err
	}
	stmt := s.update
	if tx != nil {
		stmt = tx.Stmt(stmt)
	}
	_, err = stmt.ExecContext(ctx, id, max)
	if err != nil {
		return err
	}
	return err
}

// ResetAll will reset all configurable limits to the default (no-limit).
func (s *Store) ResetAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return err
	}
	_, err = s.resetAll.ExecContext(ctx)
	return err
}

// Max will return the current max value for the given limit.
func (s *Store) Max(ctx context.Context, id ID) (int, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return 0, err
	}
	err = id.Valid()
	if err != nil {
		return 0, err
	}
	var max int
	err = s.findOne.QueryRowContext(ctx, id).Scan(&max)
	if errors.Is(err, sql.ErrNoRows) {
		return -1, nil
	}
	if err != nil {
		return 0, err
	}
	return max, nil
}

// All will get the current value of all limits.
func (s *Store) All(ctx context.Context) (Limits, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return nil, err
	}
	rows, err := s.findAll.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var id string
	var max int
	l := make(Limits, 8)
	for rows.Next() {
		err = rows.Scan(&id, &max)
		if err != nil {
			return nil, err
		}
		l[ID(id)] = max
	}
	return l, nil
}

// SetMax allows setting the max value for a limit.
func (s *Store) SetMax(ctx context.Context, id ID, max int) error {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return err
	}
	err = validate.Many(id.Valid(), validate.Range("Max", max, -1, 9000))
	if err != nil {
		return err
	}

	_, err = s.setOne.ExecContext(ctx, id, max)
	return err
}
