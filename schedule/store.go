package schedule

import (
	"context"
	"database/sql"

	"github.com/target/goalert/util/sqlutil"

	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"
)

type Store interface {
	ReadStore
	Create(context.Context, *Schedule) (*Schedule, error)
	CreateScheduleTx(context.Context, *sql.Tx, *Schedule) (*Schedule, error)
	Update(context.Context, *Schedule) error
	UpdateTx(context.Context, *sql.Tx, *Schedule) error
	Delete(context.Context, string) error
	DeleteTx(context.Context, *sql.Tx, string) error
	DeleteManyTx(context.Context, *sql.Tx, []string) error
	FindMany(context.Context, []string) ([]Schedule, error)
	FindOneForUpdate(ctx context.Context, tx *sql.Tx, id string) (*Schedule, error)

	Search(context.Context, *SearchOptions) ([]Schedule, error)
}
type ReadStore interface {
	FindAll(context.Context) ([]Schedule, error)
	FindOne(context.Context, string) (*Schedule, error)
}

type DB struct {
	db *sql.DB

	create  *sql.Stmt
	update  *sql.Stmt
	findAll *sql.Stmt
	findOne *sql.Stmt
	delete  *sql.Stmt

	findOneUp *sql.Stmt

	findMany *sql.Stmt
}

func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &DB{
		db:      db,
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
func (db *DB) FindMany(ctx context.Context, ids []string) ([]Schedule, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	err = validate.ManyUUID("ScheduleID", ids, 200)
	if err != nil {
		return nil, err
	}
	userID := permission.UserID(ctx)
	rows, err := db.findMany.QueryContext(ctx, sqlutil.UUIDArray(ids), userID)
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
func (db *DB) Create(ctx context.Context, s *Schedule) (*Schedule, error) {
	return db.CreateScheduleTx(ctx, nil, s)
}

func (db *DB) CreateScheduleTx(ctx context.Context, tx *sql.Tx, s *Schedule) (*Schedule, error) {
	n, err := s.Normalize()
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}
	stmt := db.create
	if tx != nil {
		stmt = tx.Stmt(stmt)
	}
	row := stmt.QueryRowContext(ctx, n.Name, n.Description, n.TimeZone.String())
	err = row.Scan(&n.ID)
	return n, err
}

func (db *DB) Update(ctx context.Context, s *Schedule) error {
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

	_, err = db.update.ExecContext(ctx, n.ID, n.Name, n.Description, n.TimeZone.String())
	return err
}
func (db *DB) UpdateTx(ctx context.Context, tx *sql.Tx, s *Schedule) error {
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

	_, err = tx.StmtContext(ctx, db.update).ExecContext(ctx, n.ID, n.Name, n.Description, n.TimeZone.String())
	return err
}

func (db *DB) FindAll(ctx context.Context) ([]Schedule, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	rows, err := db.findAll.QueryContext(ctx)
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

func (db *DB) FindOneForUpdate(ctx context.Context, tx *sql.Tx, id string) (*Schedule, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ScheduleID", id)
	if err != nil {
		return nil, err
	}

	row := tx.StmtContext(ctx, db.findOneUp).QueryRowContext(ctx, id)
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

func (db *DB) FindOne(ctx context.Context, id string) (*Schedule, error) {
	err := validate.UUID("ScheduleID", id)
	if err != nil {
		return nil, err
	}
	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	userID := permission.UserID(ctx)
	row := db.findOne.QueryRowContext(ctx, id, userID)
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
	if len(ids) == 0 {
		return nil
	}
	err = validate.ManyUUID("ScheduleID", ids, 50)
	if err != nil {
		return err
	}
	s := db.delete
	if tx != nil {
		s = tx.StmtContext(ctx, s)
	}
	_, err = s.ExecContext(ctx, sqlutil.UUIDArray(ids))
	return err
}
