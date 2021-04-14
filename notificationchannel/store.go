package notificationchannel

import (
	"context"
	"database/sql"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

type Store interface {
	FindAll(context.Context) ([]Channel, error)
	FindOne(context.Context, string) (*Channel, error)
	CreateTx(context.Context, *sql.Tx, *Channel) (*Channel, error)
	DeleteManyTx(context.Context, *sql.Tx, []string) error
}

type DB struct {
	db *sql.DB

	findAll    *sql.Stmt
	findOne    *sql.Stmt
	create     *sql.Stmt
	deleteMany *sql.Stmt
}

func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &DB{
		db: db,

		findAll: p.P(`
			select id, name, type, value from notification_channels
		`),
		findOne: p.P(`
			select id, name, type, value from notification_channels where id = $1
		`),
		create: p.P(`
			insert into notification_channels (id, name, type, value)
			values ($1, $2, $3, $4)
		`),
		deleteMany: p.P(`DELETE FROM notification_channels WHERE id = any($1)`),
	}, p.Err
}

func (db *DB) CreateTx(ctx context.Context, tx *sql.Tx, c *Channel) (*Channel, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	n, err := c.Normalize()
	if err != nil {
		return nil, err
	}

	_, err = tx.StmtContext(ctx, db.create).ExecContext(ctx, n.ID, n.Name, n.Type, n.Value)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (db *DB) DeleteManyTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return err
	}

	err = validate.Range("Count", len(ids), 1, 100)
	if err != nil {
		return err
	}

	del := db.deleteMany
	if tx != nil {
		tx.StmtContext(ctx, del)
	}

	_, err = del.ExecContext(ctx, sqlutil.UUIDArray(ids))
	return err
}

func (db *DB) FindOne(ctx context.Context, id string) (*Channel, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("ChannelID", id)
	if err != nil {
		return nil, err
	}

	var c Channel
	err = db.findOne.QueryRowContext(ctx, id).Scan(&c.ID, &c.Name, &c.Type, &c.Value)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (db *DB) FindAll(ctx context.Context) ([]Channel, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	rows, err := db.findAll.QueryContext(ctx)
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
