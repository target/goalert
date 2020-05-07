package integrationkey

import (
	"context"
	"database/sql"

	"github.com/target/goalert/auth/authtoken"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type Store interface {
	Authorize(ctx context.Context, tok authtoken.Token, integrationType Type) (context.Context, error)
	GetServiceID(ctx context.Context, id string, integrationType Type) (string, error)
	Create(ctx context.Context, i *IntegrationKey) (*IntegrationKey, error)
	CreateKeyTx(context.Context, *sql.Tx, *IntegrationKey) (*IntegrationKey, error)
	FindOne(ctx context.Context, id string) (*IntegrationKey, error)
	FindAllByService(ctx context.Context, id string) ([]IntegrationKey, error)
	Delete(ctx context.Context, id string) error
	DeleteTx(ctx context.Context, tx *sql.Tx, id string) error
	DeleteManyTx(ctx context.Context, tx *sql.Tx, ids []string) error
}

type DB struct {
	db *sql.DB

	getServiceID     *sql.Stmt
	create           *sql.Stmt
	findOne          *sql.Stmt
	findAllByService *sql.Stmt
	delete           *sql.Stmt
}

func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &DB{
		db: db,

		getServiceID:     p.P("SELECT service_id FROM integration_keys WHERE id = $1 AND type = $2"),
		create:           p.P("INSERT INTO integration_keys (id, name, type, service_id) VALUES ($1, $2, $3, $4)"),
		findOne:          p.P("SELECT id, name, type, service_id FROM integration_keys WHERE id = $1"),
		findAllByService: p.P("SELECT id, name, type, service_id FROM integration_keys WHERE service_id = $1"),
		delete:           p.P("DELETE FROM integration_keys WHERE id = any($1)"),
	}, p.Err
}

func (db *DB) Authorize(ctx context.Context, tok authtoken.Token, t Type) (context.Context, error) {
	var serviceID string
	var err error
	permission.SudoContext(ctx, func(c context.Context) {
		serviceID, err = db.GetServiceID(c, tok.ID.String(), t)
	})
	if err == sql.ErrNoRows {
		return ctx, validation.NewFieldError("IntegrationKeyID", "not found")
	}
	if err != nil {
		return ctx, errors.Wrap(err, "lookup serviceID")
	}
	ctx = permission.ServiceSourceContext(ctx, serviceID, &permission.SourceInfo{
		Type: permission.SourceTypeIntegrationKey,
		ID:   tok.ID.String(),
	})
	return ctx, nil
}

func (db *DB) GetServiceID(ctx context.Context, id string, t Type) (string, error) {
	err := validate.Many(
		validate.UUID("IntegrationKeyID", id),
		validate.OneOf("IntegrationType", t, TypeGrafana, TypeSite24x7, TypePrometheusAlertmanager, TypeGeneric, TypeEmail),
	)
	if err != nil {
		return "", err
	}
	err = permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
	if err != nil {
		return "", err
	}

	row := db.getServiceID.QueryRowContext(ctx, id, t)

	var serviceID string
	err = row.Scan(&serviceID)
	if err == sql.ErrNoRows {
		return "", err
	}
	if err != nil {
		return "", errors.WithMessage(err, "lookup failure")
	}

	return serviceID, nil
}

func (db *DB) Create(ctx context.Context, i *IntegrationKey) (*IntegrationKey, error) {
	return db.CreateKeyTx(ctx, nil, i)
}

func (db *DB) CreateKeyTx(ctx context.Context, tx *sql.Tx, i *IntegrationKey) (*IntegrationKey, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	n, err := i.Normalize()
	if err != nil {
		return nil, err
	}

	stmt := db.create
	if tx != nil {
		stmt = tx.Stmt(stmt)
	}

	n.ID = uuid.NewV4().String()
	_, err = stmt.ExecContext(ctx, n.ID, n.Name, n.Type, n.ServiceID)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (db *DB) Delete(ctx context.Context, id string) error {
	return db.DeleteTx(ctx, nil, id)
}
func (db *DB) DeleteTx(ctx context.Context, tx *sql.Tx, id string) error {
	return db.DeleteManyTx(ctx, tx, []string{id})
}
func (db *DB) DeleteManyTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}
	err = validate.ManyUUID("IntegrationKeyID", ids, 50)
	if err != nil {
		return err
	}

	s := db.delete
	if tx != nil {
		s = tx.Stmt(s)
	}
	_, err = s.ExecContext(ctx, sqlutil.UUIDArray(ids))
	return err
}

func (db *DB) FindOne(ctx context.Context, id string) (*IntegrationKey, error) {
	err := validate.UUID("IntegrationKeyID", id)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	row := db.findOne.QueryRowContext(ctx, id)
	var i IntegrationKey
	err = scanFrom(&i, row.Scan)
	if err != nil {
		return nil, err
	}

	return &i, nil

}

func (db *DB) FindAllByService(ctx context.Context, serviceID string) ([]IntegrationKey, error) {
	err := validate.UUID("ServiceID", serviceID)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	rows, err := db.findAllByService.QueryContext(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAllFrom(rows)
}

func scanFrom(i *IntegrationKey, f func(args ...interface{}) error) error {
	return f(&i.ID, &i.Name, &i.Type, &i.ServiceID)
}

func scanAllFrom(rows *sql.Rows) (integrationKeys []IntegrationKey, err error) {
	var i IntegrationKey
	for rows.Next() {
		err = scanFrom(&i, rows.Scan)
		if err != nil {
			return nil, err
		}
		integrationKeys = append(integrationKeys, i)
	}
	return integrationKeys, nil
}
