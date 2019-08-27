package heartbeat

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/pgtype"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"
)

// Store manages heartbeat checks and recording heartbeats.
type Store interface {
	// Heartbeat records a heartbeat for the given heartbeat ID.
	Heartbeat(context.Context, string) error

	// CreateTx creates a new heartbeat check within the transaction.
	CreateTx(context.Context, *sql.Tx, *Monitor) (*Monitor, error)

	// Delete deletes the heartbeat check with the given heartbeat ID.
	DeleteTx(context.Context, *sql.Tx, ...string) error

	// FindAllByService returns all heartbeats belonging to the given service ID.
	FindAllByService(context.Context, string) ([]Monitor, error)

	// UpdateTx updates a heartbeat's fields within the transaction.
	UpdateTx(context.Context, *sql.Tx, *Monitor) error

	// FindOneTx returns a heartbeat montior for updating.
	FindOneTx(context.Context, *sql.Tx, string) (*Monitor, error)

	// FindMany returns the heartbeat monitors with the given IDs.
	FindMany(context.Context, ...string) ([]Monitor, error)
}

var _ Store = &DB{}

// DB implements Store using Postgres as a backend.
type DB struct {
	db *sql.DB

	create     *sql.Stmt
	findAll    *sql.Stmt
	findMany   *sql.Stmt
	delete     *sql.Stmt
	update     *sql.Stmt
	getSvcID   *sql.Stmt
	findOneUpd *sql.Stmt
	heartbeat  *sql.Stmt
}

func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}
	return &DB{
		db: db,

		create: p.P(`
			insert into heartbeat_monitors (
				id, name, service_id, heartbeat_interval
			) values ($1, $2, $3, $4)
		`),
		findAll: p.P(`
			select
				id, name, service_id, heartbeat_interval, last_state, last_heartbeat
			from heartbeat_monitors
			where service_id = $1
		`),
		findMany: p.P(`
			select
				id, name, service_id, heartbeat_interval, last_state, last_heartbeat
			from heartbeat_monitors
			where id = any($1)
		`),
		findOneUpd: p.P(`
			select
				id, name, service_id, heartbeat_interval, last_state, last_heartbeat
			from heartbeat_monitors
			where id = $1
			for update
		`),
		delete: p.P(`
			delete from heartbeat_monitors
			where id = any($1)
		`),
		update: p.P(`
			update heartbeat_monitors
			set
				name = $2,
				heartbeat_interval = $3
			where id = $1
		`),
		getSvcID: p.P(`select service_id from heartbeat_monitors where id = $1`),

		heartbeat: p.P(`
			update heartbeat_monitors
			set last_heartbeat = now()
			where id = $1
		`),
	}, p.Err
}

func (db *DB) CreateTx(ctx context.Context, tx *sql.Tx, m *Monitor) (*Monitor, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}
	n, err := m.Normalize()
	if err != nil {
		return nil, err
	}
	n.ID = uuid.NewV4().String()
	n.lastState = StateInactive
	var timeout pgtype.Interval
	err = timeout.Set(n.Timeout)
	if err != nil {
		return nil, err
	}
	_, err = tx.StmtContext(ctx, db.create).ExecContext(ctx, n.ID, n.Name, n.ServiceID, &timeout)
	return n, err
}
func (db *DB) Heartbeat(ctx context.Context, id string) error {
	err := validate.UUID("MonitorID", id)
	if err != nil {
		return err
	}

	_, err = db.heartbeat.ExecContext(ctx, id)
	return err
}
func (db *DB) DeleteTx(ctx context.Context, tx *sql.Tx, ids ...string) error {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return err
	}
	err = validate.ManyUUID("MonitorID", ids, 100)
	if err != nil {
		return err
	}
	s := db.delete
	if tx != nil {
		s = tx.StmtContext(ctx, s)
	}
	_, err = s.ExecContext(ctx, pq.StringArray(ids))
	return err
}
func (db *DB) UpdateTx(ctx context.Context, tx *sql.Tx, m *Monitor) error {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return err
	}
	n, err := m.Normalize()
	if err != nil {
		return err
	}
	err = validate.Many(err,
		validate.UUID("MonitorID", n.ID),
	)
	if err != nil {
		return err
	}
	stmt := db.update
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}
	var timeout pgtype.Interval
	err = timeout.Set(n.Timeout)
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, n.ID, n.Name, &timeout)
	return err
}

func (db *DB) FindOneTx(ctx context.Context, tx *sql.Tx, id string) (*Monitor, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("ID", id)
	if err != nil {
		return nil, err
	}
	row := tx.StmtContext(ctx, db.findOneUpd).QueryRowContext(ctx, id)
	var m Monitor
	err = m.scanFrom(row.Scan)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (db *DB) FindMany(ctx context.Context, ids ...string) ([]Monitor, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}

	err = validate.ManyUUID("IDs", ids, search.MaxResults)
	if err != nil {
		return nil, err
	}
	rows, err := db.findMany.QueryContext(ctx, pq.StringArray(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []Monitor
	var m Monitor
	for rows.Next() {
		err = m.scanFrom(rows.Scan)
		if err != nil {
			return nil, err
		}
		monitors = append(monitors, m)
	}

	return monitors, nil
}

func (db *DB) FindAllByService(ctx context.Context, serviceID string) ([]Monitor, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ServiceID", serviceID)
	if err != nil {
		return nil, err
	}
	rows, err := db.findAll.QueryContext(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []Monitor
	var m Monitor
	for rows.Next() {
		err = m.scanFrom(rows.Scan)
		if err != nil {
			return nil, err
		}
		monitors = append(monitors, m)
	}

	return monitors, nil
}
