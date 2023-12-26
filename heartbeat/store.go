package heartbeat

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

// Store manages heartbeat checks and recording heartbeats.
type Store struct {
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

// NewStore creates a new Store and prepares all sql statements.
func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db: db,

		create: p.P(`
			insert into heartbeat_monitors (
				id, name, service_id, heartbeat_interval, additional_details
			) values ($1, $2, $3, $4, $5)
		`),
		findAll: p.P(`
			select
				id, name, service_id, heartbeat_interval, last_state, last_heartbeat, coalesce(additional_details, '')
			from heartbeat_monitors
			where service_id = $1
		`),
		findMany: p.P(`
			select
				id, name, service_id, heartbeat_interval, last_state, last_heartbeat, coalesce(additional_details, '')
			from heartbeat_monitors
			where id = any($1)
		`),
		findOneUpd: p.P(`
			select
				id, name, service_id, heartbeat_interval, last_state, last_heartbeat, coalesce(additional_details, '')
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
				additional_details = $4
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

// CreateTx creates a new heartbeat Monitor.
func (s *Store) CreateTx(ctx context.Context, tx *sql.Tx, m *Monitor) (*Monitor, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}

	n, err := m.Normalize()
	if err != nil {
		return nil, err
	}

	var timeout pgtype.Interval
	if err = timeout.Set(n.Timeout); err != nil {
		return nil, err
	}

	n.ID = uuid.New().String()
	n.lastState = StateInactive
	_, err = tx.StmtContext(ctx, s.create).ExecContext(ctx, n.ID, n.Name, n.ServiceID, &timeout, m.AddtionalDetails)
	if err != nil {
		return nil, err
	}

	return n, nil
}

// RecordHeartbeat records a heartbeat for the given heartbeat ID.
func (s *Store) RecordHeartbeat(ctx context.Context, id string) error {
	err := validate.UUID("MonitorID", id)
	if err != nil {
		return err
	}

	_, err = s.heartbeat.ExecContext(ctx, id)

	return err
}

// DeleteTx deletes the heartbeat check with the given ID(s).
func (s *Store) DeleteTx(ctx context.Context, tx *sql.Tx, ids ...string) error {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return err
	}

	err = validate.ManyUUID("MonitorID", ids, 100)
	if err != nil {
		return err
	}

	stmt := s.delete
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	_, err = stmt.ExecContext(ctx, sqlutil.UUIDArray(ids))

	return err
}

// UpdateTx updates a heartbeat Monitor.
func (s *Store) UpdateTx(ctx context.Context, tx *sql.Tx, m *Monitor) error {
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

	stmt := s.update
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	var timeout pgtype.Interval
	if err = timeout.Set(n.Timeout); err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx, n.ID, n.Name, &timeout, m.AddtionalDetails)

	return err
}

// FindOneTx returns a heartbeat montior for updating.
func (s *Store) FindOneTx(ctx context.Context, tx *sql.Tx, id string) (*Monitor, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("ID", id)
	if err != nil {
		return nil, err
	}

	row := tx.StmtContext(ctx, s.findOneUpd).QueryRowContext(ctx, id)

	var m Monitor
	if err = m.scanFrom(row.Scan); err != nil {
		return nil, err
	}

	return &m, nil
}

// FindMany returns the heartbeat monitors with the given IDs.
//
// The order and number of returned monitors is not guaranteed.
func (s *Store) FindMany(ctx context.Context, ids []string) ([]Monitor, error) {
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

	rows, err := s.findMany.QueryContext(ctx, sqlutil.UUIDArray(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		monitors []Monitor
		m        Monitor
	)

	for rows.Next() {
		err = m.scanFrom(rows.Scan)
		if err != nil {
			return nil, err
		}

		monitors = append(monitors, m)
	}

	return monitors, nil
}

// FindAllByService returns all heartbeats belonging to the given service ID.
func (s *Store) FindAllByService(ctx context.Context, serviceID string) ([]Monitor, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Admin)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("ServiceID", serviceID)
	if err != nil {
		return nil, err
	}

	rows, err := s.findAll.QueryContext(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		monitors []Monitor
		m        Monitor
	)

	for rows.Next() {
		err = m.scanFrom(rows.Scan)
		if err != nil {
			return nil, err
		}

		monitors = append(monitors, m)
	}

	return monitors, nil
}
