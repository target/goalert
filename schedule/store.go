package schedule

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"
)

type Store struct {
	db  *sql.DB
	usr *user.Store
}

func NewStore(ctx context.Context, db *sql.DB, usr *user.Store) (*Store, error) {
	return &Store{
		db:  db,
		usr: usr,
	}, nil
}

func (store *Store) FindManyTx(ctx context.Context, tx *sql.Tx, ids []string) ([]Schedule, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	err = validate.ManyUUID("ScheduleID", ids, 200)
	if err != nil {
		return nil, err
	}

	// Convert string IDs to UUIDs
	uuids := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		uuids[i], err = uuid.Parse(id)
		if err != nil {
			return nil, err
		}
	}

	db := gadb.New(store.db)
	if tx != nil {
		db = db.WithTx(tx)
	}

	rows, err := db.SchedFindMany(ctx, gadb.SchedFindManyParams{
		Column1: uuids,
		UserID:  permission.UserNullUUID(ctx).UUID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	result := make([]Schedule, 0, len(ids))
	for _, row := range rows {
		s := Schedule{
			ID:             row.ID.String(),
			Name:           row.Name,
			Description:    row.Description,
			isUserFavorite: row.IsFavorite,
		}

		s.TimeZone, err = util.LoadLocation(row.TimeZone)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}

	return result, nil
}

func (store *Store) FindMany(ctx context.Context, ids []string) ([]Schedule, error) {
	return store.FindManyTx(ctx, nil, ids)
}

func (store *Store) FindManyByUserID(ctx context.Context, db gadb.DBTX, userID uuid.NullUUID) ([]Schedule, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	rows, err := gadb.New(db).ScheduleFindManyByUser(ctx, userID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var result []Schedule
	for _, r := range rows {
		result = append(result, Schedule{
			ID:          r.ID.String(),
			Name:        r.Name,
			Description: r.Description,
		})
	}

	return result, nil
}

func (store *Store) Create(ctx context.Context, s *Schedule) (*Schedule, error) {
	return store.CreateScheduleTx(ctx, nil, s)
}

func (store *Store) CreateScheduleTx(ctx context.Context, tx *sql.Tx, s *Schedule) (*Schedule, error) {
	n, err := s.Normalize()
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	db := gadb.New(store.db)
	if tx != nil {
		db = db.WithTx(tx)
	}

	id, err := db.SchedCreate(ctx, gadb.SchedCreateParams{
		Name:        n.Name,
		Description: n.Description,
		TimeZone:    n.TimeZone.String(),
	})
	if err != nil {
		return nil, err
	}

	n.ID = id.String()
	return n, nil
}

func (store *Store) Update(ctx context.Context, s *Schedule) error {
	n, err := s.Normalize()
	if err != nil {
		return err
	}

	err = validate.UUID("ScheduleID", s.ID)
	if err != nil {
		return err
	}

	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(n.ID)
	if err != nil {
		return err
	}

	err = gadb.New(store.db).SchedUpdate(ctx, gadb.SchedUpdateParams{
		ID:          id,
		Name:        n.Name,
		Description: n.Description,
		TimeZone:    n.TimeZone.String(),
	})
	return err
}

func (store *Store) UpdateTx(ctx context.Context, tx *sql.Tx, s *Schedule) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}
	n, err := s.Normalize()
	if err != nil {
		return err
	}

	err = validate.UUID("ScheduleID", n.ID)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(n.ID)
	if err != nil {
		return err
	}

	err = gadb.New(store.db).WithTx(tx).SchedUpdate(ctx, gadb.SchedUpdateParams{
		ID:          id,
		Name:        n.Name,
		Description: n.Description,
		TimeZone:    n.TimeZone.String(),
	})
	return err
}

func (store *Store) FindAll(ctx context.Context) ([]Schedule, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	rows, err := gadb.New(store.db).SchedFindAll(ctx)
	if err != nil {
		return nil, err
	}

	var res []Schedule
	for _, row := range rows {
		s := Schedule{
			ID:          row.ID.String(),
			Name:        row.Name,
			Description: row.Description,
		}
		s.TimeZone, err = util.LoadLocation(row.TimeZone)
		if err != nil {
			return nil, errors.Wrap(err, "parse scanned time zone")
		}
		res = append(res, s)
	}

	return res, nil
}

func (store *Store) FindOneForUpdate(ctx context.Context, tx *sql.Tx, id string) (*Schedule, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ScheduleID", id)
	if err != nil {
		return nil, err
	}

	schedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	db := gadb.New(store.db)
	if tx != nil {
		db = db.WithTx(tx)
	}

	row, err := db.SchedFindOneForUpdate(ctx, schedID)
	if err != nil {
		return nil, err
	}

	s := Schedule{
		ID:          row.ID.String(),
		Name:        row.Name,
		Description: row.Description,
	}

	s.TimeZone, err = util.LoadLocation(row.TimeZone)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (store *Store) FindOne(ctx context.Context, id string) (*Schedule, error) {
	err := validate.UUID("ScheduleID", id)
	if err != nil {
		return nil, err
	}
	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	schedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := gadb.New(store.db).SchedFindOne(ctx, gadb.SchedFindOneParams{
		ID:     schedID,
		UserID: permission.UserNullUUID(ctx).UUID,
	})
	if err != nil {
		return nil, err
	}

	s := Schedule{
		ID:             row.ID.String(),
		Name:           row.Name,
		Description:    row.Description,
		isUserFavorite: row.IsFavorite,
	}

	s.TimeZone, err = util.LoadLocation(row.TimeZone)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (store *Store) Delete(ctx context.Context, id string) error {
	return store.DeleteTx(ctx, nil, id)
}

func (store *Store) DeleteTx(ctx context.Context, tx *sql.Tx, id string) error {
	return store.DeleteManyTx(ctx, tx, []string{id})
}

func (store *Store) DeleteManyTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	err = validate.ManyUUID("ScheduleID", ids, 50)
	if err != nil {
		return err
	}

	// Convert string IDs to UUIDs
	uuids := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		uuids[i], err = uuid.Parse(id)
		if err != nil {
			return err
		}
	}

	db := gadb.New(store.db)
	if tx != nil {
		db = db.WithTx(tx)
	}

	err = db.SchedDeleteMany(ctx, uuids)
	return err
}
