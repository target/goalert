package heartbeat

import (
	"context"
	"database/sql"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"

	uuid "github.com/satori/go.uuid"
)

// Store manages heartbeat checks and recording heartbeats.
type Store interface {
	// Heartbeat records a heartbeat for the given heartbeat ID.
	Heartbeat(context.Context, string) error

	// CreateTx creates a new heartbeat check within the transaction.
	CreateTx(context.Context, *sql.Tx, *Monitor) (*Monitor, error)

	// Delete deletes the heartbeat check with the given heartbeat ID.
	DeleteTx(context.Context, *sql.Tx, string) error

	// FindAllByService returns all heartbeats belonging to the given service ID.
	FindAllByService(context.Context, string) ([]Monitor, error)
}

var _ Store = &DB{}

// DB implements Store using Postgres as a backend.
type DB struct {
	db *sql.DB

	create   *sql.Stmt
	findAll  *sql.Stmt
	delete   *sql.Stmt
	update   *sql.Stmt
	getSvcID *sql.Stmt

	heartbeat *sql.Stmt
}

func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}
	return &DB{
		db: db,

		create: p.P(`
			insert into heartbeat_monitors (
				id, name, service_id, heartbeat_interval
			) values ($1, $2, $3, ($4||' minutes')::interval)
		`),
		findAll: p.P(`
			select
				id, name, extract(epoch from heartbeat_interval)/60, last_state, trunc(extract(epoch from now()-last_heartbeat)/60)::int
			from heartbeat_monitors
			where service_id = $1
		`),
		delete: p.P(`
			delete from heartbeat_monitors
			where id = $1
		`),
		update: p.P(`
			update heartbeat_monitors
			set
				name = $2,
				heartbeat_interval = ($3||' minutes')::interval
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

	_, err = tx.Stmt(db.create).ExecContext(ctx, n.ID, n.Name, n.ServiceID, n.IntervalMinutes)
	n.lastState = StateInactive
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
func (db *DB) DeleteTx(ctx context.Context, tx *sql.Tx, id string) error {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return err
	}
	err = validate.UUID("MonitorID", id)
	if err != nil {
		return err
	}
	s := db.delete
	if tx != nil {
		s = tx.Stmt(s)
	}
	_, err = s.ExecContext(ctx, id)
	return err
}
func (db *DB) Update(ctx context.Context, m *Monitor) error {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return err
	}
	n, err := m.Normalize()
	err = validate.Many(err,
		validate.UUID("MonitorID", n.ID),
	)
	if err != nil {
		return err
	}
	_, err = db.update.ExecContext(ctx, n.ID, n.Name, n.IntervalMinutes)
	return err
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
	for rows.Next() {
		var m Monitor
		m.ServiceID = serviceID
		err = rows.Scan(&m.ID, &m.Name, &m.IntervalMinutes, &m.lastState, &m.lastHeartbeatMinutes)
		if err != nil {
			return nil, err
		}
		monitors = append(monitors, m)
	}

	return monitors, nil
}
