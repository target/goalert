package calendarsubscription

import (
	"context"
	"database/sql"
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
	"time"
)

type Store struct {
	db *sql.DB
	findOne *sql.Stmt
	create  *sql.Stmt
	update  *sql.Stmt
	delete  *sql.Stmt
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db: db,
		findOne: p.P(`
				SELECT
					cs.id,
					cs.name,
					cs.user_id,
					cs.last_access,
					cs.disabled
				FROM calendar_subscriptions cs
				WHERE cs.id = $1
			`),
		create: p.P(`
				INSERT INTO calendar_subscriptions (id, name, user_id, last_access, disabled)
				VALUES ($1, $2, $3, $4, $5)
			`),
		update: p.P(`UPDATE calendar_subscriptions SET name = $2, disabled = $3 WHERE id = $1`),
		delete: p.P(`DELETE FROM calendar_subscriptions WHERE id = any($1)`),
	}, p.Err
}

func (b *Store) FindOne(ctx context.Context, id string) (*CalendarSubscription, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("CalendarSubscriptionID", id)
	if err != nil {
		return nil, err
	}

	var cs CalendarSubscription
	err = b.findOne.QueryRowContext(ctx, id).Scan(&cs.ID, &cs.Name, &cs.UserID, &cs.LastAccess, &cs.Disabled)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

func (b *Store) CreateSubscriptionTx(ctx context.Context, tx *sql.Tx, cs *CalendarSubscription) (res bool, err error) {
	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return false, err
	}

	n, err := cs.Normalize()
	if err != nil {
		return false, err
	}

	stmt := b.create
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	n.ID = uuid.NewV4().String()
	n.UserID = permission.UserID(ctx)
	n.LastAccess = time.Now()
	n.Disabled = false

	_, err = stmt.ExecContext(ctx, n.ID, n.Name, n.UserID, n.LastAccess, n.Disabled)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *Store) UpdateSubscriptionTx(ctx context.Context, tx *sql.Tx, cs *CalendarSubscription) (err error) {
	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	err = validate.UUID("CalendarSubscriptionID", cs.ID)
	if err != nil {
		return err
	}

	n, err := cs.Normalize()
	if err != nil {
		return err
	}

	stmt := b.update
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	_, err = stmt.ExecContext(ctx, n.ID, n.Name, n.Disabled)
	if err != nil {
		return err
	}

	return err
}

func (b *Store) DeleteSubscriptionsTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	err = validate.ManyUUID("CalendarSubscriptionID", ids, 50)
	if err != nil {
		return err
	}

	cs := b.delete
	if tx != nil {
		cs = tx.StmtContext(ctx, cs)
	}
	_, err = cs.ExecContext(ctx, sqlutil.UUIDArray(ids))
	return err
}
