package notificationrule

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

// Store allows the lookup and management of NotificationRules.
type Store struct {
	db *sql.DB

	insert       *sql.Stmt
	delete       *sql.Stmt
	findAll      *sql.Stmt
	lookupUserID *sql.Stmt
}

// NewDB will create a DB backend from a sql.DB. An error will be returned if statements fail to prepare.
func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	prep := &util.Prepare{DB: db, Ctx: ctx}
	p := prep.P
	s := &Store{db: db}

	s.insert = p("INSERT INTO user_notification_rules (id,user_id,delay_minutes,contact_method_id) VALUES ($1,$2,$3,$4)")
	s.findAll = p("SELECT id,user_id,delay_minutes,contact_method_id FROM user_notification_rules WHERE user_id = $1")
	s.delete = p("DELETE FROM user_notification_rules WHERE id = any($1)")
	s.lookupUserID = p("SELECT user_id FROM user_notification_rules WHERE id = any($1)")

	return s, prep.Err
}

// Insert implements the NotificationRuleStore interface by inserting the new NotificationRule into the database.
// A new ID is always created.
func (s *Store) Insert(ctx context.Context, n *NotificationRule) (*NotificationRule, error) {
	return s.CreateTx(ctx, nil, n)
}

// CreateTx implements the NotificationRuleStore interface by inserting the new NotificationRule into the database.
// A new ID is always created.
func (s *Store) CreateTx(ctx context.Context, tx *sql.Tx, n *NotificationRule) (*NotificationRule, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.MatchUser(n.UserID))
	if err != nil {
		return nil, err
	}

	n, err = n.Normalize(false)
	if err != nil {
		return nil, err
	}

	n.ID = uuid.New().String()

	_, err = wrapTx(ctx, tx, s.insert).ExecContext(ctx, n.ID, n.UserID, n.DelayMinutes, n.ContactMethodID)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func wrapTx(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt) *sql.Stmt {
	if tx == nil {
		return stmt
	}

	return tx.StmtContext(ctx, stmt)
}

// DeleteTx will delete notification rules with the provided ids.
func (s *Store) DeleteTx(ctx context.Context, tx *sql.Tx, ids ...string) error {
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
		_, err = wrapTx(ctx, tx, s.delete).ExecContext(ctx, sqlutil.UUIDArray(ids))
		return err
	}

	rows, err := wrapTx(ctx, tx, s.lookupUserID).QueryContext(ctx, sqlutil.UUIDArray(ids))
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
	_, err = wrapTx(ctx, tx, s.delete).ExecContext(ctx, sqlutil.UUIDArray(ids))
	return err
}

// FindAll implements the NotificationRuleStore interface.
func (s *Store) FindAll(ctx context.Context, userID string) ([]NotificationRule, error) {
	err := validate.UUID("UserID", userID)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.System, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}

	rows, err := s.findAll.QueryContext(ctx, userID)
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
