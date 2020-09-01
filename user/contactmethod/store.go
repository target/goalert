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

// Store allows the lookup and management of ContactMethods.
type Store interface {
	Insert(context.Context, *ContactMethod) (*ContactMethod, error)
	CreateTx(context.Context, *sql.Tx, *ContactMethod) (*ContactMethod, error)
	Update(context.Context, *ContactMethod) error
	UpdateTx(context.Context, *sql.Tx, *ContactMethod) error
	Delete(ctx context.Context, id string) error
	FindOne(ctx context.Context, id string) (*ContactMethod, error)
	FindOneTx(ctx context.Context, tx *sql.Tx, id string) (*ContactMethod, error)
	FindMany(ctx context.Context, ids []string) ([]ContactMethod, error)
	FindAll(ctx context.Context, userID string) ([]ContactMethod, error)
	DeleteTx(ctx context.Context, tx *sql.Tx, id ...string) error
	EnableByValue(context.Context, Type, string) error
	DisableByValue(context.Context, Type, string) error

	MetadataByTypeValue(ctx context.Context, tx *sql.Tx, t Type, value string) (*Metadata, error)
	SetCarrierV1MetadataByTypeValue(ctx context.Context, tx *sql.Tx, t Type, value string, m *Metadata) error
}

// DB implements the ContactMethodStore against a *sql.DB backend.
type DB struct {
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

// NewDB will create a DB backend from a sql.DB. An error will be returned if statements fail to prepare.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}
	return &DB{
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
			INSERT INTO user_contact_methods (id,name,type,value,disabled,user_id)
			VALUES ($1,$2,$3,$4,$5,$6)
		`),
		findOne: p.P(`
			SELECT id,name,type,value,disabled,user_id
			FROM user_contact_methods
			WHERE id = $1
		`),
		findOneUpd: p.P(`
			SELECT id,name,type,value,disabled,user_id
			FROM user_contact_methods
			WHERE id = $1
			FOR UPDATE
		`),
		findMany: p.P(`
			SELECT id,name,type,value,disabled,user_id
			FROM user_contact_methods
			WHERE id = any($1)
		`),
		findAll: p.P(`
			SELECT id,name,type,value,disabled,user_id
			FROM user_contact_methods
			WHERE user_id = $1
		`),
		update: p.P(`
				UPDATE user_contact_methods
				SET name = $2, disabled = $3
				WHERE id = $1
			`),
		delete: p.P(`
				DELETE FROM user_contact_methods
				WHERE id = any($1)
			`),
	}, p.Err
}

func (db *DB) MetadataByTypeValue(ctx context.Context, tx *sql.Tx, typ Type, value string) (*Metadata, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}
	var data json.RawMessage
	var t time.Time
	err = wrapTx(ctx, tx, db.metaTV).QueryRowContext(ctx, typ, value).Scan(&data, &t)
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

func (db *DB) SetCarrierV1MetadataByTypeValue(ctx context.Context, tx *sql.Tx, typ Type, value string, newM *Metadata) error {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return err
	}
	var ownTx bool
	if tx == nil {
		tx, err = db.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		ownTx = true
	}
	m, err := db.MetadataByTypeValue(ctx, tx, typ, value)
	if err != nil {
		return err
	}
	m.CarrierV1 = newM.CarrierV1
	m.CarrierV1.UpdatedAt = m.FetchedAt

	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	_, err = tx.StmtContext(ctx, db.setMetaTV).ExecContext(ctx, typ, value, data)
	if err != nil {
		return err
	}

	if ownTx {
		return tx.Commit()
	}

	return nil
}

func (db *DB) EnableByValue(ctx context.Context, t Type, v string) error {
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
	err = db.enable.QueryRowContext(ctx, n.Type, n.Value).Scan(&id)

	if err == nil {
		// NOTE: maintain a record of consent/dissent
		logCtx := log.WithFields(ctx, log.Fields{
			"contactMethodID": id,
		})

		log.Logf(logCtx, "Contact method START code received.")
	}

	return err
}

func (db *DB) DisableByValue(ctx context.Context, t Type, v string) error {
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
	err = db.disable.QueryRowContext(ctx, n.Type, n.Value).Scan(&id)

	if err == nil {
		// NOTE: maintain a record of consent/dissent
		logCtx := log.WithFields(ctx, log.Fields{
			"contactMethodID": id,
		})

		log.Logf(logCtx, "Contact method STOP code received.")
	}

	return err
}

// Insert implements the ContactMethodStore interface by inserting the new ContactMethod into the database.
// A new ID is always created.
func (db *DB) Insert(ctx context.Context, c *ContactMethod) (*ContactMethod, error) {
	return db.CreateTx(ctx, nil, c)
}

// CreateTx implements the ContactMethodStore interface by inserting the new ContactMethod into the database.
// A new ID is always created.
func (db *DB) CreateTx(ctx context.Context, tx *sql.Tx, c *ContactMethod) (*ContactMethod, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.MatchUser(c.UserID))
	if err != nil {
		return nil, err
	}

	n, err := c.Normalize()
	if err != nil {
		return nil, err
	}

	_, err = wrapTx(ctx, tx, db.insert).ExecContext(ctx, n.ID, n.Name, n.Type, n.Value, n.Disabled, n.UserID)
	if err != nil {
		return nil, err
	}

	return n, nil
}

// Delete implements the ContactMethodStore interface.
func (db *DB) Delete(ctx context.Context, id string) error {
	return db.DeleteTx(ctx, nil, id)
}

func wrapTx(ctx context.Context, tx *sql.Tx, stmt *sql.Stmt) *sql.Stmt {
	if tx == nil {
		return stmt
	}

	return tx.StmtContext(ctx, stmt)
}

// DeleteTx implements the ContactMethodStore interface.
func (db *DB) DeleteTx(ctx context.Context, tx *sql.Tx, ids ...string) error {
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

// FindOneTx implements the ContactMethodStore interface.
func (db *DB) FindOneTx(ctx context.Context, tx *sql.Tx, id string) (*ContactMethod, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ContactMethodID", id)
	if err != nil {
		return nil, err
	}

	var c ContactMethod
	row := wrapTx(ctx, tx, db.findOneUpd).QueryRowContext(ctx, id)
	err = row.Scan(&c.ID, &c.Name, &c.Type, &c.Value, &c.Disabled, &c.UserID)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// FindOne implements the ContactMethodStore interface.
func (db *DB) FindOne(ctx context.Context, id string) (*ContactMethod, error) {
	err := validate.UUID("ContactMethodID", id)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	var c ContactMethod
	row := db.findOne.QueryRowContext(ctx, id)
	err = row.Scan(&c.ID, &c.Name, &c.Type, &c.Value, &c.Disabled, &c.UserID)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Update implements the ContactMethodStore interface.
func (db *DB) Update(ctx context.Context, c *ContactMethod) error {
	return db.UpdateTx(ctx, nil, c)
}

// UpdateTx implements the ContactMethodStore interface.
func (db *DB) UpdateTx(ctx context.Context, tx *sql.Tx, c *ContactMethod) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	n, err := c.Normalize()
	if err != nil {
		return err
	}

	cm, err := db.FindOneTx(ctx, tx, c.ID)
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
		_, err = wrapTx(ctx, tx, db.update).ExecContext(ctx, n.ID, n.Name, n.Disabled)
		return err
	}

	err = permission.LimitCheckAny(ctx, permission.MatchUser(cm.UserID))
	if err != nil {
		return err
	}

	_, err = wrapTx(ctx, tx, db.update).ExecContext(ctx, n.ID, n.Name, n.Disabled)
	return err
}

// FindMany will fetch all contact methods matching the given ids.
func (db *DB) FindMany(ctx context.Context, ids []string) ([]ContactMethod, error) {
	err := validate.ManyUUID("ContactMethodID", ids, 50)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	rows, err := db.findMany.QueryContext(ctx, sqlutil.UUIDArray(ids))
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
		err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.Value, &c.Disabled, &c.UserID)
		if err != nil {
			return nil, err
		}

		contactMethods = append(contactMethods, c)
	}
	return contactMethods, nil
}

// FindAll implements the ContactMethodStore interface.
func (db *DB) FindAll(ctx context.Context, userID string) ([]ContactMethod, error) {
	err := validate.UUID("UserID", userID)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	rows, err := db.findAll.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAll(rows)
}
