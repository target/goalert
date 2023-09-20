package integrationkey

import (
	"context"
	"database/sql"

	"github.com/target/goalert/auth/authtoken"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Store struct {
	db *sql.DB
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db: db,
	}, p.Err
}

func (s *Store) queries(tx *sql.Tx) *gadb.Queries {
	if tx != nil {
		return gadb.New(tx)
	}
	return gadb.New(s.db)
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
	keyUUID, err := validate.ParseUUID("IntegrationKeyID", id)
	err = validate.Many(
		err,
		validate.OneOf("IntegrationType", t, TypeGrafana, TypeSite24x7, TypePrometheusAlertmanager, TypeGeneric, TypeEmail),
	)
	if err != nil {
		return "", err
	}
	err = permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
	if err != nil {
		return "", err
	}

	serviceID, err := gadb.New(s.db).GetServiceID(ctx, gadb.GetServiceIDParams{
		ID:   keyUUID,
		Type: gadb.EnumIntegrationKeysType(t),
	})

	if errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if err != nil {
		return "", errors.WithMessage(err, "lookup failure")
	}

	return serviceID.String(), nil
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

	serviceUUID, err := uuid.Parse(n.ServiceID)
	if err != nil {
		return nil, err
	}

	keyUUID := uuid.New()
	n.ID = keyUUID.String()
	err = s.queries(tx).CreateIntegrationKey(ctx, gadb.CreateIntegrationKeyParams{
		ID:        keyUUID,
		Name:      n.Name,
		Type:      gadb.EnumIntegrationKeysType(n.Type),
		ServiceID: serviceUUID,
	})
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

	uuids, err := validate.ParseManyUUID("IntegrationKeyID", ids, 50)
	if err != nil {
		return err
	}

	err = s.queries(tx).DeleteIntegrationKey(ctx, uuids)
	return err
}

func (s *Store) FindOne(ctx context.Context, id string) (*IntegrationKey, error) {
	keyUUID, err := validate.ParseUUID("IntegrationKeyID", id)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	row, err := gadb.New(s.db).FindOneIntegrationKey(ctx, keyUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &IntegrationKey{
		ID:        row.ID.String(),
		Name:      row.Name,
		Type:      Type(row.Type),
		ServiceID: row.ServiceID.String(),
	}, nil
}

func (s *Store) FindAllByService(ctx context.Context, serviceID string) ([]IntegrationKey, error) {
	serviceUUID, err := validate.ParseUUID("ServiceID", serviceID)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	rows, err := gadb.New(s.db).FindIntegrationKeysByService(ctx, serviceUUID)
	if err != nil {
		return nil, err
	}
	keys := make([]IntegrationKey, len(rows))
	for i, row := range rows {
		keys[i] = IntegrationKey{
			ID:        row.ID.String(),
			Name:      row.Name,
			Type:      Type(row.Type),
			ServiceID: row.ServiceID.String(),
		}
	}
	return keys, nil
}
