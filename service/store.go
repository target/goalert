package service

import (
	"context"
	"database/sql"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation/validate"

	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	return &Store{db: db}, nil
}

func (s *Store) FindOneForUpdate(ctx context.Context, tx *sql.Tx, id string) (*Service, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ServiceID", id)
	if err != nil {
		return nil, err
	}

	res, err := gadb.New(tx).ServiceFindOneForUpdate(ctx, uuid.MustParse(id))
	if err != nil {
		return nil, err
	}

	return &Service{
		ID:                   res.ID.String(),
		Name:                 res.Name,
		Description:          res.Description,
		EscalationPolicyID:   res.EscalationPolicyID.String(),
		MaintenanceExpiresAt: res.MaintenanceExpiresAt.Time,
	}, nil
}

// FindMany returns slice of Service objects given a slice of serviceIDs
func (s *Store) FindMany(ctx context.Context, ids []string) ([]Service, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	err = validate.ManyUUID("ServiceIDs", ids, 100)
	if err != nil {
		return nil, err
	}

	uuids := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		uuids[i] = uuid.MustParse(id)
	}

	res, err := gadb.New(s.db).ServiceFindMany(ctx, gadb.ServiceFindManyParams{ServiceIds: uuids, UserID: uuid.MustParse(permission.UserID(ctx))})
	if err != nil {
		return nil, err
	}

	svcs := make([]Service, len(res))
	for i, r := range res {
		svcs[i] = Service{
			ID:                   r.ID.String(),
			Name:                 r.Name,
			Description:          r.Description,
			EscalationPolicyID:   r.EscalationPolicyID.String(),
			MaintenanceExpiresAt: r.MaintenanceExpiresAt.Time,
			isUserFavorite:       r.IsUserFavorite,
		}
	}

	return svcs, nil
}

func (s *Store) CreateServiceTx(ctx context.Context, tx *sql.Tx, svc *Service) (*Service, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	n, err := svc.Normalize()
	if err != nil {
		return nil, err
	}

	err = gadb.New(tx).ServiceInsert(ctx, gadb.ServiceInsertParams{
		ID:                 uuid.MustParse(n.ID),
		Name:               n.Name,
		Description:        n.Description,
		EscalationPolicyID: uuid.MustParse(n.EscalationPolicyID),
	})
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (s *Store) DeleteManyTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}
	err = validate.ManyUUID("ServiceID", ids, 50)
	if err != nil {
		return err
	}

	uuids := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		uuids[i] = uuid.MustParse(id)
	}

	return gadb.New(tx).ServiceDeleteMany(ctx, uuids)
}

func (s *Store) UpdateTx(ctx context.Context, tx *sql.Tx, svc *Service) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	n, err := svc.Normalize()
	if err != nil {
		return err
	}

	err = validate.UUID("ServiceID", n.ID)
	if err != nil {
		return err
	}

	return gadb.New(tx).ServiceUpdate(ctx, gadb.ServiceUpdateParams{
		ID:                 uuid.MustParse(n.ID),
		Name:               n.Name,
		Description:        n.Description,
		EscalationPolicyID: uuid.MustParse(n.EscalationPolicyID),
		MaintenanceExpiresAt: sql.NullTime{
			Time:  n.MaintenanceExpiresAt,
			Valid: !n.MaintenanceExpiresAt.IsZero(),
		},
	})
}

func (s *Store) FindOne(ctx context.Context, id string) (*Service, error) {
	res, err := s.FindMany(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, sql.ErrNoRows
	}
	return &res[0], nil
}

func (s *Store) FindAllByEP(ctx context.Context, epID string) ([]Service, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	res, err := gadb.New(s.db).ServiceFindManyByEP(ctx, gadb.ServiceFindManyByEPParams{EscalationPolicyID: uuid.MustParse(epID), UserID: uuid.MustParse(permission.UserID(ctx))})
	if err != nil {
		return nil, err
	}

	svcs := make([]Service, len(res))
	for i, r := range res {
		svcs[i] = Service{
			ID:                   r.ID.String(),
			Name:                 r.Name,
			Description:          r.Description,
			EscalationPolicyID:   r.EscalationPolicyID.String(),
			MaintenanceExpiresAt: r.MaintenanceExpiresAt.Time,
			isUserFavorite:       r.IsUserFavorite,
		}
	}

	return svcs, nil
}
