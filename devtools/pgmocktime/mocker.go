package pgmocktime

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Mocker struct {
	db     *pgxpool.Pool
	dbName string

	schema string
	err    error
}

var mx sync.Mutex

// New creates a new Mocker capable of manipulating time in a postgres database.
func New(ctx context.Context, dbURL string) (*Mocker, error) {
	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	var dbName string
	err = db.QueryRow(ctx, `SELECT current_database()`).Scan(&dbName)
	if err != nil {
		return nil, fmt.Errorf("select current time: %w", err)
	}

	return &Mocker{db: db, dbName: dbName, schema: "pgmocktime"}, nil
}

func (m *Mocker) exec(ctx context.Context, queryFormat string, args ...interface{}) {
	if m.err != nil {
		log.Println("skipping", queryFormat, args)
		return
	}

	query := fmt.Sprintf(queryFormat, args...)
	_, err := m.db.Exec(ctx, query)
	if err != nil {
		m.err = fmt.Errorf("exec: %s: %w", query, err)
	}
}

func (m *Mocker) safeSchema() string {
	return pgx.Identifier{m.schema}.Sanitize()
}

func (m *Mocker) safeDB() string {
	return pgx.Identifier{m.dbName}.Sanitize()
}

func (m *Mocker) readErr(action string) (err error) {
	err = m.err
	m.err = nil
	if err != nil {
		return fmt.Errorf("%s: %w", action, err)
	}
	return nil
}

// Inject instruments the database for the manipulation of time.
func (m *Mocker) Inject(ctx context.Context) error {
	m.exec(ctx, `create schema if not exists %s`, m.safeSchema())

	m.exec(ctx, `
	create unlogged table if not exists %s.flux_capacitor (
		ok BOOL PRIMARY KEY,
		ref_time TIMESTAMPTZ NOT NULL DEFAULT transaction_timestamp(),
		base_time TIMESTAMPTZ NOT NULL DEFAULT transaction_timestamp(),
		speed FLOAT NOT NULL DEFAULT 1.0,
		CHECK(ok)
	)`, m.safeSchema())

	m.exec(ctx, `insert into %s.flux_capacitor (ok) values (true) on conflict do nothing`, m.safeSchema())

	mx.Lock()
	m.exec(ctx, `
	create or replace function %s.now()
	RETURNS timestamptz
	AS $$
		DECLARE
			_ref_time timestamptz;
			_base_time timestamptz;
			_speed FLOAT;
		BEGIN
			SELECT ref_time, base_time, speed INTO _ref_time, _base_time, _speed FROM %s.flux_capacitor;
			RETURN _base_time + (current_timestamp - _ref_time) * _speed;
		end;
	$$ language plpgsql;
	`,
		m.safeSchema(),
		m.safeSchema(),
	)
	mx.Unlock()

	m.exec(ctx, `alter database %s set search_path = "$user", public, %s, pg_catalog`, m.safeDB(), m.safeSchema())
	if m.err != nil {
		log.Println("skipping", m.err)
		return m.readErr("inject")
	}

	// update all columns from `pg_catalog.now()` to `schema.now()`
	rows, err := m.db.Query(ctx, `select table_name, column_name from information_schema.columns where column_default = 'pg_catalog.now()' or column_default = 'now()'`)
	if err != nil {
		return fmt.Errorf("update columns: %w", err)
	}
	defer rows.Close()

	type col struct {
		table string
		name  string
	}
	var cols []col

	for rows.Next() {
		var c col
		if err := rows.Scan(&c.table, &c.name); err != nil {
			return fmt.Errorf("scan columns: %w", err)
		}
		cols = append(cols, c)
	}

	for _, c := range cols {
		m.exec(ctx, `alter table %s alter column %s set default %s.now()`,
			pgx.Identifier{c.table}.Sanitize(),
			pgx.Identifier{c.name}.Sanitize(),
			m.safeSchema(),
		)
	}

	return m.readErr("inject")
}

// Remove removes instrumentation from the database.
func (m *Mocker) Remove(ctx context.Context) error {
	m.exec(ctx, `alter database %s reset search_path`, m.safeDB())
	m.exec(ctx, `drop schema if exists %s cascade`, m.safeSchema())
	return m.readErr("remove")
}

// Close closes the database connection.
func (m *Mocker) Close() error { m.db.Close(); return nil }

// AdvanceTime advances the time by the given duration.
func (m *Mocker) AdvanceTime(ctx context.Context, d time.Duration) error {
	m.exec(ctx, `update %s.flux_capacitor set ref_time = current_timestamp, base_time = %s.now() + '%d hours'::interval + '%d milliseconds'::interval`, m.safeSchema(), m.safeSchema(), d/time.Hour, (d % time.Hour).Milliseconds())
	return m.readErr("advance time")
}

// SetTime sets the database time to the given time.
func (m *Mocker) SetTime(ctx context.Context, t time.Time) error {
	m.exec(ctx, `update %s.flux_capacitor set ref_time = current_timestamp, base_time = '%s'::timestamptz`, m.safeSchema(), t.Format(time.RFC3339Nano))
	return m.readErr("set time")
}

// SetSpeed sets the speed of time, 1.0 is the normal flow of time.
func (m *Mocker) SetSpeed(ctx context.Context, speed float64) error {
	m.exec(ctx, `update %s.flux_capacitor set speed = %f, ref_time = current_timestamp, base_time = %s.now()`, m.safeSchema(), speed, m.safeSchema())
	return m.readErr("set speed")
}

// Reset sets the time to the current real time and resumes the normal flow of time.
func (m *Mocker) Reset(ctx context.Context) error {
	m.exec(ctx, `update %s.flux_capacitor set speed = 1.0, ref_time = current_timestamp, base_time = current_timestamp`, m.safeSchema())
	return m.readErr("reset")
}
