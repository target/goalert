package integrationkey

import (
	"context"
	"database/sql"

	"github.com/target/goalert/auth/authtoken"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Store struct {
	db *sql.DB

	getServiceID     *sql.Stmt
	create           *sql.Stmt
	findOne          *sql.Stmt
	findAllByService *sql.Stmt
	delete           *sql.Stmt
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db: db,

		getServiceID:     p.P("SELECT service_id FROM integration_keys WHERE id = $1 AND type = $2"),
		create:           p.P("INSERT INTO integration_keys (id, name, type, service_id) VALUES ($1, $2, $3, $4)"),
		findOne:          p.P("SELECT id, name, type, service_id FROM integration_keys WHERE id = $1"),
		findAllByService: p.P("SELECT id, name, type, service_id FROM integration_keys WHERE service_id = $1"),
		delete:           p.P("DELETE FROM integration_keys WHERE id = any($1)"),
	}, p.Err
}

func (s *Store) Authorize(ctx context.Context, tok authtoken.Token, t Type) (context.Context, error) {
	var serviceID string
	var err error
	permission.SudoContext(ctx, func(c context.Context) {
		serviceID, err = s.GetServiceID(c, tok.ID.String(), t)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return ctx, permission.Unauthorized()
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

func (s *Store) GetServiceID(ctx context.Context, id string, t Type) (string, error) {
	err := validate.Many(
		validate.UUID("IntegrationKeyID", id),
		validate.OneOf("IntegrationType", t, TypeGrafana, TypeSite24x7, TypePrometheusAlertmanager, TypeGeneric, TypeNotify, TypeEmail),
	)
	if err != nil {
		return "", err
	}
	err = permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
	if err != nil {
		return "", err
	}

	row := s.getServiceID.QueryRowContext(ctx, id, t)

	var serviceID string
	err = row.Scan(&serviceID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if err != nil {
		return "", errors.WithMessage(err, "lookup failure")
	}

	return serviceID, nil
}

func (s *Store) Create(ctx context.Context, i *IntegrationKey) (*IntegrationKey, error) {
	return s.CreateKeyTx(ctx, nil, i)
}

func (s *Store) CreateKeyTx(ctx context.Context, tx *sql.Tx, i *IntegrationKey) (*IntegrationKey, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	n, err := i.Normalize()
	if err != nil {
		return nil, err
	}

	stmt := s.create
	if tx != nil {
		stmt = tx.Stmt(stmt)
	}

	n.ID = uuid.New().String()
	_, err = stmt.ExecContext(ctx, n.ID, n.Name, n.Type, n.ServiceID)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	return s.DeleteTx(ctx, nil, id)
}

func (s *Store) DeleteTx(ctx context.Context, tx *sql.Tx, id string) error {
	return s.DeleteManyTx(ctx, tx, []string{id})
}

func (s *Store) DeleteManyTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}
	err = validate.ManyUUID("IntegrationKeyID", ids, 50)
	if err != nil {
		return err
	}

	stmt := s.delete
	if tx != nil {
		stmt = tx.Stmt(stmt)
	}
	_, err = stmt.ExecContext(ctx, sqlutil.UUIDArray(ids))
	return err
}

func (s *Store) FindOne(ctx context.Context, id string) (*IntegrationKey, error) {
	err := validate.UUID("IntegrationKeyID", id)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	row := s.findOne.QueryRowContext(ctx, id)
	var i IntegrationKey
	err = scanFrom(&i, row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &i, nil
}

func (s *Store) FindAllByService(ctx context.Context, serviceID string) ([]IntegrationKey, error) {
	err := validate.UUID("ServiceID", serviceID)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	rows, err := s.findAllByService.QueryContext(ctx, serviceID)
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
