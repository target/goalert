package override

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// Store is used to manage active overrides.
type Store struct {
	db *sql.DB

	findUO    *sql.Stmt
	createUO  *sql.Stmt
	deleteUO  *sql.Stmt
	findAllUO *sql.Stmt
	updateUO  *sql.Stmt

	lock *sql.Stmt

	findUOUpdate *sql.Stmt
}

// NewStore initializes a new DB using an existing sql connection.
func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db: db,

		lock: p.P(`LOCK user_overrides IN EXCLUSIVE MODE`),

		findUOUpdate: p.P(`
		select
			id,
			add_user_id, 
			remove_user_id,
			start_time,
			end_time,
			tgt_schedule_id
		from user_overrides
		where id = $1
		for update
	`),
		findUO: p.P(`
			select
				id,
				add_user_id, 
				remove_user_id,
				start_time,
				end_time,
				tgt_schedule_id
			from user_overrides
			where id = $1
		`),
		updateUO: p.P(`
			update user_overrides
			set
				add_user_id = $2,
				remove_user_id = $3,
				start_time = $4,
				end_time = $5,
				tgt_schedule_id = $6
			where id = $1
		`),
		createUO: p.P(`
			insert into user_overrides (
				id,
				add_user_id,
				remove_user_id,
				start_time,
				end_time,
				tgt_schedule_id
			) values ($1, $2, $3, $4, $5, $6)`),
		deleteUO: p.P(`delete from user_overrides where id = any($1)`),
		findAllUO: p.P(`
			select
				id,
				add_user_id, 
				remove_user_id,
				start_time,
				end_time
			from user_overrides
			where
				tgt_schedule_id = $1 and
				(start_time, end_time) OVERLAPS ($2, $3)
		`),
	}, p.Err
}

func (s *Store) withTx(ctx context.Context, tx *sql.Tx, fn func(tx *sql.Tx) error) error {
	if tx == nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		// Since this is a helper method, we don't
		// have much context to work with.
		defer sqlutil.Rollback(ctx, "override", tx)

		err = s.withTx(ctx, tx, fn)
		if err != nil {
			return err
		}
		return tx.Commit()
	}

	_, err := tx.StmtContext(ctx, s.lock).ExecContext(ctx)
	if err != nil {
		return err
	}

	return fn(tx)
}

func (s *Store) execContext(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt, args ...interface{}) error {
	return s.withTx(ctx, tx, func(tx *sql.Tx) error {
		_, err := tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
		return err
	})
}

func (s *Store) FindOneUserOverrideTx(ctx context.Context, tx *sql.Tx, id string, forUpdate bool) (*UserOverride, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("OverrideID", id)
	if err != nil {
		return nil, err
	}

	var o UserOverride
	var add, rem, schedTgt sql.NullString
	err = s.withTx(ctx, tx, func(tx *sql.Tx) error {
		var row *sql.Row
		if forUpdate {
			row = tx.StmtContext(ctx, s.findUOUpdate).QueryRowContext(ctx, id)
		} else {
			row = tx.StmtContext(ctx, s.findUO).QueryRowContext(ctx, id)
		}

		return row.Scan(&o.ID, &add, &rem, &o.Start, &o.End, &schedTgt)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	o.AddUserID = add.String
	o.RemoveUserID = rem.String
	if schedTgt.Valid {
		o.Target = assignment.ScheduleTarget(schedTgt.String)
	}

	return &o, nil
}

// UpdateUserOverrideTx updates an existing UserOverride, inside an optional transaction.
func (s *Store) UpdateUserOverrideTx(ctx context.Context, tx *sql.Tx, o *UserOverride) error {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return err
	}
	n, err := o.Normalize()
	if err != nil {
		return err
	}
	err = validate.UUID("ID", n.ID)
	if err != nil {
		return err
	}
	if !n.End.After(time.Now()) {
		return validation.NewFieldError("End", "must be in the future")
	}
	var add, rem sql.NullString
	if n.AddUserID != "" {
		add.Valid = true
		add.String = n.AddUserID
	}
	if n.RemoveUserID != "" {
		rem.Valid = true
		rem.String = n.RemoveUserID
	}
	var schedTgt sql.NullString
	if n.Target.TargetType() == assignment.TargetTypeSchedule {
		schedTgt.Valid = true
		schedTgt.String = n.Target.TargetID()
	}
	return s.execContext(ctx, tx, s.updateUO, n.ID, add, rem, n.Start, n.End, schedTgt)
}

// UpdateUserOverride updates an existing UserOverride.
func (s *Store) UpdateUserOverride(ctx context.Context, o *UserOverride) error {
	return s.UpdateUserOverrideTx(ctx, nil, o)
}

// CreateUserOverrideTx adds a UserOverride to the DB with a new ID.
func (s *Store) CreateUserOverrideTx(ctx context.Context, tx *sql.Tx, o *UserOverride) (*UserOverride, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}
	n, err := o.Normalize()
	if err != nil {
		return nil, err
	}
	if !n.End.After(time.Now()) {
		return nil, validation.NewFieldError("End", "must be in the future")
	}
	n.ID = uuid.New().String()
	var add, rem sql.NullString
	if n.AddUserID != "" {
		add.Valid = true
		add.String = n.AddUserID
	}
	if n.RemoveUserID != "" {
		rem.Valid = true
		rem.String = n.RemoveUserID
	}
	var schedTgt sql.NullString
	if n.Target.TargetType() == assignment.TargetTypeSchedule {
		schedTgt.Valid = true
		schedTgt.String = n.Target.TargetID()
	}
	err = s.execContext(ctx, tx, s.createUO, n.ID, add, rem, n.Start, n.End, schedTgt)
	if err != nil {
		return nil, err
	}

	return n, nil
}

// DeleteUserOverride removes a UserOverride from the DB matching the given ID.
func (s *Store) DeleteUserOverrideTx(ctx context.Context, tx *sql.Tx, ids ...string) error {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	err = validate.ManyUUID("UserOverrideID", ids, 50)
	if err != nil {
		return err
	}

	return s.execContext(ctx, tx, s.deleteUO, sqlutil.UUIDArray(ids))
}

// FindAllUserOverrides will return all UserOverrides that belong to the provided Target within the provided time range.
func (s *Store) FindAllUserOverrides(ctx context.Context, start, end time.Time, t assignment.Target) ([]UserOverride, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}
	err = validate.Many(
		validate.OneOf("TargetType", t.TargetType(), assignment.TargetTypeSchedule),
		validate.UUID("TargetID", t.TargetID()),
	)
	if err != nil {
		return nil, err
	}

	var schedTgt sql.NullString
	if t.TargetType() == assignment.TargetTypeSchedule {
		schedTgt.Valid = true
		schedTgt.String = t.TargetID()
	}

	rows, err := s.findAllUO.QueryContext(ctx, schedTgt, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []UserOverride
	var o UserOverride
	var add, rem sql.NullString
	o.Target = t
	for rows.Next() {
		err = rows.Scan(&o.ID, &add, &rem, &o.Start, &o.End)
		if err != nil {
			return nil, err
		}
		// no need to check `Valid` since we're find with the empty string
		o.AddUserID = add.String
		o.RemoveUserID = rem.String
		result = append(result, o)
	}

	return result, nil
}
