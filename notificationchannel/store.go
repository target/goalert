package notificationchannel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

type Store struct {
	db *sql.DB

	findAll    *sql.Stmt
	findOne    *sql.Stmt
	findMany   *sql.Stmt
	create     *sql.Stmt
	deleteMany *sql.Stmt

	updateName  *sql.Stmt
	findByValue *sql.Stmt
	lock        *sql.Stmt
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db: db,

		findAll: p.P(`
			select id, name, type, value from notification_channels
		`),
		findOne: p.P(`
			select id, name, type, value from notification_channels where id = $1
		`),
		findMany: p.P(`
			select id, name, type, value from notification_channels where id = any($1)
		`),
		create: p.P(`
			insert into notification_channels (id, name, type, value)
			values ($1, $2, $3, $4)
		`),
		updateName: p.P(`update notification_channels set name = $2 where id = $1`),
		deleteMany: p.P(`DELETE FROM notification_channels WHERE id = any($1)`),

		findByValue: p.P(`select id, name from notification_channels where type = $1 and value = $2`),

		// Lock the table so only one tx can insert/update at a time, but allows the above SELECT FOR UPDATE to run
		// so only required changes block.
		lock: p.P(`LOCK notification_channels IN SHARE ROW EXCLUSIVE MODE`),
	}, p.Err
}

func stmt(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt) *sql.Stmt {
	if tx == nil {
		return stmt
	}

	return tx.StmtContext(ctx, stmt)
}

func (s *Store) FindMany(ctx context.Context, ids []string) ([]Channel, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	err = validate.ManyUUID("ID", ids, search.MaxResults)
	if err != nil {
		return nil, err
	}

	rows, err := s.findMany.QueryContext(ctx, sqlutil.UUIDArray(ids))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []Channel
	for rows.Next() {
		var c Channel
		err = rows.Scan(&c.ID, &c.Name, &c.Type, &c.Value)
		if err != nil {
			return nil, err
		}

		channels = append(channels, c)
	}

	return channels, nil
}

func (s *Store) MapToID(ctx context.Context, tx *sql.Tx, c *Channel) (uuid.UUID, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return uuid.UUID{}, err
	}

	n, err := c.Normalize()
	if err != nil {
		return uuid.UUID{}, err
	}

	var id sqlutil.NullUUID
	var name sql.NullString
	err = stmt(ctx, tx, s.findByValue).QueryRowContext(ctx, n.Type, n.Value).Scan(&id, &name)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("lookup existing entry: %w", err)
	}

	if id.Valid && name.String == c.Name {
		// short-circuit if it already exists and is up-to-date.
		return id.UUID, nil
	}

	var ownTx bool
	if tx == nil {
		ownTx = true
		tx, err = s.db.BeginTx(ctx, nil)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("start tx: %w", err)
		}
		defer tx.Rollback()
	}

	_, err = tx.StmtContext(ctx, s.lock).ExecContext(ctx)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("acquire lock: %w", err)
	}

	// try again after exclusive lock
	err = tx.StmtContext(ctx, s.findByValue).QueryRowContext(ctx, n.Type, n.Value).Scan(&id, &name)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("lookup existing entry exclusively: %w", err)
	}
	if id.Valid && name.String == c.Name {
		// short-circuit if it already exists and is up-to-date.
		return id.UUID, nil
	}
	if !id.Valid {
		// create new one
		id.Valid = true
		id.UUID = uuid.New()
		_, err = tx.StmtContext(ctx, s.create).ExecContext(ctx, id, n.Name, n.Type, n.Value)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("create new NC: %w", err)
		}
	} else {
		// update existing name
		_, err = tx.StmtContext(ctx, s.updateName).ExecContext(ctx, id, n.Name)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("update NC name: %w", err)
		}
	}

	if ownTx {
		return id.UUID, tx.Commit()
	}
	return id.UUID, nil
}

func (s *Store) DeleteManyTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return err
	}

	err = validate.Range("Count", len(ids), 1, 100)
	if err != nil {
		return err
	}

	del := s.deleteMany
	if tx != nil {
		tx.StmtContext(ctx, del)
	}

	_, err = del.ExecContext(ctx, sqlutil.UUIDArray(ids))
	return err
}

func (s *Store) FindOne(ctx context.Context, id uuid.UUID) (*Channel, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	var c Channel
	err = s.findOne.QueryRowContext(ctx, id).Scan(&c.ID, &c.Name, &c.Type, &c.Value)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) FindAll(ctx context.Context) ([]Channel, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	rows, err := s.findAll.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []Channel
	for rows.Next() {
		var c Channel
		err = rows.Scan(&c.ID, &c.Name, &c.Type, &c.Value)
		if err != nil {
			return nil, err
		}
		channels = append(channels, c)
	}

	return channels, nil
}
