package contactmethod

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// Store implements the lookup and management of ContactMethods against a *sql.Store backend.
type Store struct {
	db *sql.DB

	insert       *sql.Stmt
	update       *sql.Stmt
	delete       *sql.Stmt
	findOne      *sql.Stmt
	findOneUpd   *sql.Stmt
	findMany     *sql.Stmt
	findAll      *sql.Stmt
	lookupUserID *sql.Stmt
	enable       *sql.Stmt
	disable      *sql.Stmt
	metaTV       *sql.Stmt
	setMetaTV    *sql.Stmt
	now          *sql.Stmt
}

// NewStore will create a DB backend from a sql.DB. An error will be returned if statements fail to prepare.
func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}
	return &Store{
		db: db,

		now: p.P(`select now()`),

		metaTV: p.P(`
			SELECT coalesce(metadata, '{}'), now()
			FROM user_contact_methods
			WHERE type = $1 AND value = $2
		`),
		setMetaTV: p.P(`
			UPDATE user_contact_methods
			SET metadata = $3
			WHERE type = $1 AND value = $2
		`),

		enable: p.P(`
			UPDATE user_contact_methods
			SET disabled = false
			WHERE type = $1
				AND value = $2
			RETURNING id
		`),
		disable: p.P(`
			UPDATE user_contact_methods
			SET disabled = true
			WHERE type = $1
				AND value = $2
			RETURNING id
		`),
		lookupUserID: p.P(`
			SELECT DISTINCT user_id
			FROM user_contact_methods
			WHERE id = any($1)
		`),
		insert: p.P(`
			INSERT INTO user_contact_methods (id,name,type,value,disabled,user_id,enable_status_updates)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`),
		findOne: p.P(`
			SELECT id,name,type,value,disabled,user_id,last_test_verify_at,enable_status_updates,pending
			FROM user_contact_methods
			WHERE id = $1
		`),
		findOneUpd: p.P(`
			SELECT id,name,type,value,disabled,user_id,last_test_verify_at,enable_status_updates,pending
			FROM user_contact_methods
			WHERE id = $1
			FOR UPDATE
		`),
		findMany: p.P(`
			SELECT id,name,type,value,disabled,user_id,last_test_verify_at,enable_status_updates,pending
			FROM user_contact_methods
			WHERE id = any($1)
		`),
		findAll: p.P(`
			SELECT id,name,type,value,disabled,user_id,last_test_verify_at,enable_status_updates,pending
			FROM user_contact_methods
			WHERE user_id = $1
		`),
		update: p.P(`
				UPDATE user_contact_methods
				SET name = $2, disabled = $3, enable_status_updates = $4
				WHERE id = $1
			`),
		delete: p.P(`
				DELETE FROM user_contact_methods
				WHERE id = any($1)
			`),
	}, p.Err
}

func (s *Store) MetadataByTypeValue(ctx context.Context, tx *sql.Tx, typ Type, value string) (*Metadata, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}
	var data json.RawMessage
	var t time.Time
	err = wrapTx(ctx, tx, s.metaTV).QueryRowContext(ctx, typ, value).Scan(&data, &t)
	if err != nil {
		return nil, err
	}

	var m Metadata
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	m.FetchedAt = t

	return &m, nil
}

func (s *Store) SetCarrierV1MetadataByTypeValue(ctx context.Context, tx *sql.Tx, typ Type, value string, newM *Metadata) error {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return err
	}
	var ownTx bool
	if tx == nil {
		tx, err = s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer sqlutil.Rollback(ctx, "cm: set carrier metadata", tx)

		ownTx = true
	}
	m, err := s.MetadataByTypeValue(ctx, tx, typ, value)
	if err != nil {
		return err
	}
	m.CarrierV1 = newM.CarrierV1
	m.CarrierV1.UpdatedAt = m.FetchedAt

	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	_, err = tx.StmtContext(ctx, s.setMetaTV).ExecContext(ctx, typ, value, data)
	if err != nil {
		return err
	}

	if ownTx {
		return tx.Commit()
	}

	return nil
}

func (s *Store) EnableByValue(ctx context.Context, t Type, v string) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}

	c := ContactMethod{Name: "Enable", Type: t, Value: v}
	n, err := c.Normalize()
	if err != nil {
		return err
	}

	var id string
	err = s.enable.QueryRowContext(ctx, n.Type, n.Value).Scan(&id)

	if err == nil {
		// NOTE: maintain a record of consent/dissent
		logCtx := log.WithFields(ctx, log.Fields{
			"contactMethodID": id,
		})

		log.Logf(logCtx, "Contact method START code received.")
	}

	return err
}

func (s *Store) DisableByValue(ctx context.Context, t Type, v string) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}

	c := ContactMethod{Name: "Disable", Type: t, Value: v}
	n, err := c.Normalize()
	if err != nil {
		return err
	}

	var id string
	err = s.disable.QueryRowContext(ctx, n.Type, n.Value).Scan(&id)

	if err == nil {
		// NOTE: maintain a record of consent/dissent
		logCtx := log.WithFields(ctx, log.Fields{
			"contactMethodID": id,
		})

		log.Logf(logCtx, "Contact method STOP code received.")
	}

	return err
}

// CreateTx inserts the new ContactMethod into the database. A new ID is always created.
func (s *Store) CreateTx(ctx context.Context, tx *sql.Tx, c *ContactMethod) (*ContactMethod, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.MatchUser(c.UserID))
	if err != nil {
		return nil, err
	}

	n, err := c.Normalize()
	if err != nil {
		return nil, err
	}

	_, err = wrapTx(ctx, tx, s.insert).ExecContext(ctx, n.ID, n.Name, n.Type, n.Value, n.Disabled, n.UserID, n.StatusUpdates)
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

// Delete removes the ContactMethod from the database using the provided ID within a transaction.
func (s *Store) DeleteTx(ctx context.Context, tx *sql.Tx, ids ...string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	err = validate.ManyUUID("ContactMethodID", ids, 50)
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

// FindOneTx finds the contact method from the database using the provided ID within a transaction.
func (s *Store) FindOneTx(ctx context.Context, tx *sql.Tx, id string) (*ContactMethod, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ContactMethodID", id)
	if err != nil {
		return nil, err
	}

	var c ContactMethod
	row := wrapTx(ctx, tx, s.findOneUpd).QueryRowContext(ctx, id)
	err = row.Scan(&c.ID, &c.Name, &c.Type, &c.Value, &c.Disabled, &c.UserID, &c.lastTestVerifyAt, &c.StatusUpdates, &c.Pending)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// FindOne finds the contact method from the database using the provided ID.
func (s *Store) FindOne(ctx context.Context, id string) (*ContactMethod, error) {
	err := validate.UUID("ContactMethodID", id)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	var c ContactMethod
	row := s.findOne.QueryRowContext(ctx, id)
	err = row.Scan(&c.ID, &c.Name, &c.Type, &c.Value, &c.Disabled, &c.UserID, &c.lastTestVerifyAt, &c.StatusUpdates, &c.Pending)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateTx updates the contact method with the newly provided values within a transaction.
func (s *Store) UpdateTx(ctx context.Context, tx *sql.Tx, c *ContactMethod) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	n, err := c.Normalize()
	if err != nil {
		return err
	}

	cm, err := s.FindOneTx(ctx, tx, c.ID)
	if err != nil {
		return err
	}
	if n.Type != cm.Type {
		return validation.NewFieldError("Type", "cannot update type of contact method")
	}
	if n.Value != cm.Value {
		return validation.NewFieldError("Value", "cannot update value of contact method")
	}
	if n.UserID != cm.UserID {
		return validation.NewFieldError("UserID", "cannot update owner of contact method")
	}

	if permission.Admin(ctx) {
		_, err = wrapTx(ctx, tx, s.update).ExecContext(ctx, n.ID, n.Name, n.Disabled, n.StatusUpdates)
		return err
	}

	err = permission.LimitCheckAny(ctx, permission.MatchUser(cm.UserID))
	if err != nil {
		return err
	}

	_, err = wrapTx(ctx, tx, s.update).ExecContext(ctx, n.ID, n.Name, n.Disabled, n.StatusUpdates)
	return err
}

// FindMany will fetch all contact methods matching the given ids.
func (s *Store) FindMany(ctx context.Context, ids []string) ([]ContactMethod, error) {
	err := validate.ManyUUID("ContactMethodID", ids, 50)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	rows, err := s.findMany.QueryContext(ctx, sqlutil.UUIDArray(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAll(rows)
}

func scanAll(rows *sql.Rows) ([]ContactMethod, error) {
	var contactMethods []ContactMethod
	for rows.Next() {
		var c ContactMethod
		err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.Value, &c.Disabled, &c.UserID, &c.lastTestVerifyAt, &c.StatusUpdates, &c.Pending)
		if err != nil {
			return nil, err
		}

		contactMethods = append(contactMethods, c)
	}
	return contactMethods, nil
}

// FindAll finds all contact methods from the database associated with the given user ID.
func (s *Store) FindAll(ctx context.Context, userID string) ([]ContactMethod, error) {
	err := validate.UUID("UserID", userID)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	rows, err := s.findAll.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAll(rows)
}
