package calendarsubscription

import (
	"context"
	"database/sql"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"
)

type Store interface {
	FindOne(context.Context, string) (*CalendarSubscription, error)
}

type DB struct {
	db *sql.DB

	findOne *sql.Stmt
}

func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &DB{
		db:      db,
		findOne:  p.P(`
			SELECT
				cs.id,
				cs.name,
				cs.user_id,
				cs.last_access,
				cs.disabled
			FROM calendar_subscriptions cs
			WHERE cs.id = $1
		`),
	}, p.Err
}

func (db *DB) FindOne(ctx context.Context, id string) (*CalendarSubscription, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("CalendarSubscriptionID", id)
	if err != nil {
		return nil, err
	}

	var cs CalendarSubscription
	err = db.findOne.QueryRowContext(ctx, id).Scan(&cs.ID, &cs.Name, &cs.UserID, &cs.LastAccess, &cs.Disabled)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}
