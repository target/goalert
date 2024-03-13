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
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

type Store struct {
	db *sql.DB

	create  *sql.Stmt
	update  *sql.Stmt
	findAll *sql.Stmt
	findOne *sql.Stmt
	delete  *sql.Stmt

	findData    *sql.Stmt
	findUpdData *sql.Stmt
	updateData  *sql.Stmt
	insertData  *sql.Stmt

	findOneUp *sql.Stmt

	findMany *sql.Stmt

	usr *user.Store
}

func NewStore(ctx context.Context, db *sql.DB, usr *user.Store) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db:  db,
		usr: usr,

		findData:    p.P(`SELECT data FROM schedule_data WHERE schedule_id = $1`),
		findUpdData: p.P(`SELECT data FROM schedule_data WHERE schedule_id = $1 FOR UPDATE`),
		insertData:  p.P(`INSERT INTO schedule_data (schedule_id, data) VALUES ($1, '{}')`),
		updateData:  p.P(`UPDATE schedule_data SET data = $2 WHERE schedule_id = $1`),

		create:  p.P(`INSERT INTO schedules (id, name, description, time_zone) VALUES (DEFAULT, $1, $2, $3) RETURNING id`),
		update:  p.P(`UPDATE schedules SET name = $2, description = $3, time_zone = $4 WHERE id = $1`),
		findAll: p.P(`SELECT id, name, description, time_zone FROM schedules`),
		findOne: p.P(`
			SELECT
				s.id,
				s.name,
				s.description,
				s.time_zone,
				fav IS DISTINCT FROM NULL
			FROM schedules s
			LEFT JOIN user_favorites fav ON
				fav.tgt_schedule_id = s.id AND fav.user_id = $2
			WHERE s.id = $1
		`),
		findOneUp: p.P(`SELECT id, name, description, time_zone FROM schedules WHERE id = $1 FOR UPDATE`),

		findMany: p.P(`
			SELECT
				s.id,
				s.name,
				s.description,
				s.time_zone,
				fav is distinct from null
			FROM schedules s
			LEFT JOIN user_favorites fav ON
				fav.tgt_schedule_id = s.id AND fav.user_id = $2
			WHERE s.id = any($1)
		`),

		delete: p.P(`DELETE FROM schedules WHERE id = any($1)`),
	}, p.Err
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

	stmt := store.findMany
	if tx != nil {
		stmt = tx.Stmt(stmt)
	}
	rows, err := stmt.QueryContext(ctx, sqlutil.UUIDArray(ids), permission.UserNullUUID(ctx))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Schedule, 0, len(ids))
	var s Schedule
	var tz string
	for rows.Next() {
		err = rows.Scan(&s.ID, &s.Name, &s.Description, &tz, &s.isUserFavorite)
		if err != nil {
			return nil, err
		}

		s.TimeZone, err = util.LoadLocation(tz)
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
	stmt := store.create
	if tx != nil {
		stmt = tx.Stmt(stmt)
	}
	row := stmt.QueryRowContext(ctx, n.Name, n.Description, n.TimeZone.String())
	err = row.Scan(&n.ID)
	return n, err
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

	_, err = store.update.ExecContext(ctx, n.ID, n.Name, n.Description, n.TimeZone.String())
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

	_, err = tx.StmtContext(ctx, store.update).ExecContext(ctx, n.ID, n.Name, n.Description, n.TimeZone.String())
	return err
}

func (store *Store) FindAll(ctx context.Context) ([]Schedule, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	rows, err := store.findAll.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var s Schedule
	var tz string
	var res []Schedule
	for rows.Next() {
		err = rows.Scan(&s.ID, &s.Name, &s.Description, &tz)
		if err != nil {
			return nil, err
		}
		s.TimeZone, err = util.LoadLocation(tz)
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

	row := tx.StmtContext(ctx, store.findOneUp).QueryRowContext(ctx, id)
	var s Schedule
	var tz string
	err = row.Scan(&s.ID, &s.Name, &s.Description, &tz)
	if err != nil {
		return nil, err
	}

	s.TimeZone, err = util.LoadLocation(tz)
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

	row := store.findOne.QueryRowContext(ctx, id, permission.UserNullUUID(ctx))
	var s Schedule
	var tz string
	err = row.Scan(&s.ID, &s.Name, &s.Description, &tz, &s.isUserFavorite)
	if err != nil {
		return nil, err
	}

	s.TimeZone, err = util.LoadLocation(tz)
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
	s := store.delete
	if tx != nil {
		s = tx.StmtContext(ctx, s)
	}
	_, err = s.ExecContext(ctx, sqlutil.UUIDArray(ids))
	return err
}
