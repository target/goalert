package notificationrule

import (
	"context"
	"database/sql"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"

	"github.com/target/goalert/validation/validate"

	uuid "github.com/satori/go.uuid"
)

// Store allows the lookup and management of NotificationRules.
type Store interface {
	Insert(context.Context, *NotificationRule) (*NotificationRule, error)
	UpdateDelay(ctx context.Context, id string, delay int) error
	Delete(ctx context.Context, id string) error
	DeleteTx(ctx context.Context, tx *sql.Tx, ids ...string) error
	CreateTx(context.Context, *sql.Tx, *NotificationRule) (*NotificationRule, error)
	FindOne(ctx context.Context, id string) (*NotificationRule, error)
	FindAll(ctx context.Context, userID string) ([]NotificationRule, error)

	WrapTx(*sql.Tx) Store
	DoTx(func(Store) error) error
}

// DB implements the NotificationRuleStore against a *sql.DB backend.
type DB struct {
	db *sql.DB

	insert       *sql.Stmt
	update       *sql.Stmt
	delete       *sql.Stmt
	findOne      *sql.Stmt
	findAll      *sql.Stmt
	lookupUserID *sql.Stmt
}

// NewDB will create a DB backend from a sql.DB. An error will be returned if statements fail to prepare.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	prep := &util.Prepare{DB: db, Ctx: ctx}
	p := prep.P
	s := &DB{db: db}

	s.insert = p("INSERT INTO user_notification_rules (id,user_id,delay_minutes,contact_method_id) VALUES ($1,$2,$3,$4)")
	s.findOne = p("SELECT id,user_id,delay_minutes,contact_method_id FROM user_notification_rules WHERE id = $1 LIMIT 1")
	s.findAll = p("SELECT id,user_id,delay_minutes,contact_method_id FROM user_notification_rules WHERE user_id = $1")
	s.update = p("UPDATE user_notification_rules SET delay_minutes = $2 WHERE id = $1")
	s.delete = p("DELETE FROM user_notification_rules WHERE id = any($1)")
	s.lookupUserID = p("SELECT user_id FROM user_notification_rules WHERE id = any($1)")

	return s, prep.Err
}

// WrapTx will wrap the NotificationRuleDB for use within the given transaction.
func (db *DB) WrapTx(tx *sql.Tx) Store {
	return &DB{
		insert:  tx.Stmt(db.insert),
		findOne: tx.Stmt(db.findOne),
		findAll: tx.Stmt(db.findAll),
		update:  tx.Stmt(db.update),
		delete:  tx.Stmt(db.delete),
	}
}

// DoTx will perform a transaction with the NotificationRuleStore.
func (d *DB) DoTx(f func(Store) error) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = f(d.WrapTx(tx))
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Insert implements the NotificationRuleStore interface by inserting the new NotificationRule into the database.
// A new ID is always created.
func (db *DB) Insert(ctx context.Context, n *NotificationRule) (*NotificationRule, error) {
	return db.CreateTx(ctx, nil, n)
}

// CreateTx implements the NotificationRuleStore interface by inserting the new NotificationRule into the database.
// A new ID is always created.
func (db *DB) CreateTx(ctx context.Context, tx *sql.Tx, n *NotificationRule) (*NotificationRule, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.MatchUser(n.UserID))
	if err != nil {
		return nil, err
	}

	n, err = n.Normalize(false)
	if err != nil {
		return nil, err
	}

	n.ID = uuid.NewV4().String()

	_, err = wrapTx(ctx, tx, db.insert).ExecContext(ctx, n.ID, n.UserID, n.DelayMinutes, n.ContactMethodID)
	if err != nil {
		return nil, err
	}

	return n, nil
}

// Delete implements the NotificationRuleStore interface.
func (db *DB) Delete(ctx context.Context, id string) error {
	return db.DeleteTx(ctx, nil, id)
}

func wrapTx(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt) *sql.Stmt {
	if tx == nil {
		return stmt
	}

	return tx.StmtContext(ctx, stmt)
}

// DeleteTx will delete notification rules with the provided ids.
func (db *DB) DeleteTx(ctx context.Context, tx *sql.Tx, ids ...string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	err = validate.ManyUUID("NotificationRuleID", ids, 50)
	if err != nil {
		return err
	}

	if permission.Admin(ctx) {
		_, err = wrapTx(ctx, tx, db.delete).ExecContext(ctx, sqlutil.UUIDArray(ids))
		return err
	}

	rows, err := wrapTx(ctx, tx, db.lookupUserID).QueryContext(ctx, sqlutil.UUIDArray(ids))
	if err != nil {
		return err
	}
	defer rows.Close()

	var checks []permission.Checker
	var userID string
	for rows.Next() {
		err = rows.Scan(&userID)
		if err != nil {
			return err
		}
		checks = append(checks, permission.MatchUser(userID))
	}

	err = permission.LimitCheckAny(ctx, checks...)
	if err != nil {
		return err
	}
	_, err = wrapTx(ctx, tx, db.delete).ExecContext(ctx, sqlutil.UUIDArray(ids))
	return err
}

// FindOne implements the NotificationRuleStore interface.
func (db *DB) FindOne(ctx context.Context, id string) (*NotificationRule, error) {
	err := validate.UUID("NotificationRuleID", id)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}
	var n NotificationRule
	row := db.findOne.QueryRowContext(ctx, id)
	err = row.Scan(&n.ID, &n.UserID, &n.DelayMinutes, &n.ContactMethodID)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// Update implements the NotificationRuleStore interface.
func (db *DB) UpdateDelay(ctx context.Context, id string, delay int) error {
	err := validate.UUID("NotificationRuleID", id)
	if err != nil {
		return err
	}
	err = validateDelay(delay)
	if err != nil {
		return err
	}

	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	if permission.Admin(ctx) {
		_, err = db.update.ExecContext(ctx, id, delay)
		return err
	}

	var userID string

	row := db.lookupUserID.QueryRowContext(ctx, sqlutil.UUIDArray{id})
	err = row.Scan(&userID)
	if err != nil {
		return err
	}

	err = permission.LimitCheckAny(ctx, permission.MatchUser(userID))
	if err != nil {
		return err
	}

	_, err = db.update.ExecContext(ctx, id, delay)
	return err
}

// FindAll implements the NotificationRuleStore interface.
func (db *DB) FindAll(ctx context.Context, userID string) ([]NotificationRule, error) {
	err := validate.UUID("UserID", userID)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.System, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}

	rows, err := db.findAll.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notificationrules := []NotificationRule{}
	for rows.Next() {
		var n NotificationRule
		err = rows.Scan(&n.ID, &n.UserID, &n.DelayMinutes, &n.ContactMethodID)
		if err != nil {
			return nil, err
		}
		notificationrules = append(notificationrules, n)
	}

	return notificationrules, nil
}
