package calendarsubscription

import (
	"context"
	"database/sql"
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

type Store struct {
	db *sql.DB
	findOne *sql.Stmt
	create  *sql.Stmt
	update  *sql.Stmt
	delete  *sql.Stmt
	findAll *sql.Stmt
	findOneUpd *sql.Stmt
}

type Config struct {
	NotificationMinutes []int `json:"notification_minutes"`
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db: db,
		findOne: p.P(`SELECT * FROM user_calendar_subscriptions cs WHERE cs.id = $1`),
		create: p.P(`
				INSERT INTO user_calendar_subscriptions (user_id, id, name, config, schedule_id)
				VALUES ($1, $2, $3, $4, $5)
			`),
		update: p.P(`UPDATE user_calendar_subscriptions SET name = $3, disabled = $4, config = $5 WHERE id = $1 AND user_id = $2`),
		delete: p.P(`DELETE FROM user_calendar_subscriptions WHERE id = any($1) AND user_id = $2`),
		findAll: p.P(`SELECT * FROM user_calendar_subscriptions WHERE user_id = $1`),
		findOneUpd: p.P(`SELECT id, name, user_id, disabled, config FROM user_calendar_subscriptions WHERE id = $1 FOR UPDATE`),
	}, p.Err
}


func wrapTx(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt) *sql.Stmt {
	if tx == nil {
		return stmt
	}
	return tx.StmtContext(ctx, stmt)
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
	err = b.findOne.QueryRowContext(ctx, id).Scan(&cs.ID, &cs.Name, &cs.UserID, &cs.LastAccess, &cs.Disabled, &cs.ScheduleID, &cs.Config)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

func (b *Store) CreateSubscriptionTx(ctx context.Context, tx *sql.Tx, cs *CalendarSubscription) (*CalendarSubscription, error) {
	cs.UserID = permission.UserID(ctx)
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.MatchUser(cs.UserID))
	if err != nil {
		return nil, err
	}
	cs.ID = uuid.NewV4().String()
	cs, err = cs.Normalize()
	if err != nil {
		return nil, err
	}

	_, err = wrapTx(ctx, tx, b.create).ExecContext(ctx, cs.UserID, cs.ID, cs.Name, cs.Config, cs.ScheduleID)
	if err != nil {
		return nil, err
	}
	return cs, nil
}
func (b *Store) FindOneForUpdateTx(ctx context.Context, tx *sql.Tx, id string) (*CalendarSubscription, error) {
	var userID = permission.UserID(ctx)
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.MatchUser(userID))
	if err != nil {
		return nil, err
	}
	err = validate.UUID("CalendarSubscriptionID", id)
	if err != nil {
		return nil, err
	}

	var cs CalendarSubscription

	row := wrapTx(ctx, tx, b.findOneUpd).QueryRowContext(ctx, id)
	err = row.Scan(&cs.ID, &cs.Name, &cs.UserID, &cs.Disabled, &cs.Config)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}
func (b *Store) UpdateSubscriptionTx(ctx context.Context, tx *sql.Tx, cs *CalendarSubscription) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.MatchUser(cs.UserID))
	if err != nil {
		return err
	}

	n, err := cs.Normalize()
	if err != nil {
		return err
	}

	update, err := b.FindOneForUpdateTx(ctx, tx, cs.ID)
	if err != nil {
		return err
	}
	if n.ScheduleID != update.ScheduleID {
		return validation.NewFieldError("ScheduleID", "cannot update schedule id of calendar subscription")
	}
	if n.LastAccess != update.LastAccess {
		return validation.NewFieldError("Last Access", "cannot update last access of calendar subscription")
	}
	if n.UserID != update.UserID {
		return validation.NewFieldError("UserID", "cannot update owner of calendar subscription")
	}

	_, err = wrapTx(ctx, tx, b.update).ExecContext(ctx, cs.ID, cs.UserID, cs.Name, cs.Disabled, cs.Config)

	return err
}
func (b *Store) FindAll(ctx context.Context) ([]CalendarSubscription, error) {
	var userID = permission.UserID(ctx)
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.MatchUser(userID))
	if err != nil {
		return nil, err
	}

	rows, err := b.findAll.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	calendarsubscriptions := []CalendarSubscription{}
	for rows.Next() {
		var cs CalendarSubscription
		err = rows.Scan(&cs.ID, &cs.Name, &cs.UserID, &cs.LastAccess, &cs.Disabled, &cs.ScheduleID, &cs.Config)
		if err != nil {
			return nil, err
		}
		calendarsubscriptions = append(calendarsubscriptions, cs)
	}

	return calendarsubscriptions, nil
}

func (b *Store) DeleteSubscriptionsTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.MatchUser(permission.UserID(ctx)))
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

	_, err = wrapTx(ctx, tx, b.delete).ExecContext(ctx, sqlutil.UUIDArray(ids), permission.UserID(ctx))
	return err
}
