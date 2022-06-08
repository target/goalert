package pgmocktime

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
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
	db, err := pgxpool.Connect(context.Background(), dbURL)
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

func (m *Mocker) setSpeed(ctx context.Context, s float64) {
	if m.err != nil {
		return
	}

	mx.Lock()
	defer mx.Unlock()

	m.exec(ctx, `
	create or replace function %s.time_speed()
	returns float
	as $$
		begin
		return %f;
		end;
	$$ language plpgsql;
	`,
		m.safeSchema(),
		s,
	)
}

func (m *Mocker) setOffset(ctx context.Context, offset time.Duration) {
	if m.err != nil {
		return
	}

	mx.Lock()
	defer mx.Unlock()
	m.exec(ctx, `
	create or replace function %s.time_offset()
	returns interval
	as $$
		begin
		return '%d milliseconds'::interval;
		end;
	$$ language plpgsql;
	`,
		m.safeSchema(),
		offset/time.Millisecond,
	)
}

func (m *Mocker) timestamp(ctx context.Context) time.Time {
	if m.err != nil {
		return time.Time{}
	}

	var t time.Time
	m.err = m.db.QueryRow(ctx, `SELECT current_timestamp`).Scan(&t)
	if m.err != nil {
		m.err = fmt.Errorf("select current time: %w", m.err)
	}

	return t
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
	m.setOffset(ctx, 0)
	m.setSpeed(ctx, 1.0)

	refTime := m.timestamp(ctx)
	if refTime.IsZero() {
		return m.readErr("inject")
	}

	mx.Lock()
	m.exec(ctx, `
	create or replace function %s.now()
	returns timestamptz
	as $$
		begin
		return '%s'::timestamptz + (current_timestamp - '%s'::timestamptz) * time_speed() + time_offset();
		end;
	$$ language plpgsql;
	`,
		m.safeSchema(),
		refTime.Format(time.RFC3339Nano),
		refTime.Format(time.RFC3339Nano),
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
	var offset time.Duration
	err := m.db.QueryRow(ctx, fmt.Sprintf(`SELECT %s.time_offset()`, m.safeSchema())).Scan(&offset)
	if err != nil {
		return fmt.Errorf("select time offset: %w", err)
	}

	m.setOffset(ctx, offset+d)
	return m.readErr("advance time")
}

// SetOffset sets the offset of time to the given absolute duration compared to the current real-world time.
func (m *Mocker) SetOffset(ctx context.Context, offset time.Duration) error {
	m.setOffset(ctx, offset)
	return m.readErr("set offset")
}

// SetTime sets the database time to the given time.
func (m *Mocker) SetTime(ctx context.Context, t time.Time) error {
	var curTime time.Time
	var offset time.Duration
	err := m.db.QueryRow(ctx, fmt.Sprintf(`SELECT %s.time_offset(), %s.now()`, m.safeSchema(), m.safeSchema())).Scan(&offset, &t)
	if err != nil {
		return fmt.Errorf("select time offset: %w", err)
	}

	offset += t.Sub(curTime)
	m.setOffset(ctx, offset)
	return m.readErr("set time")
}

// SetSpeed sets the speed of time, 1.0 is the normal flow of time.
func (m *Mocker) SetSpeed(ctx context.Context, speed float64) error {
	m.setSpeed(ctx, speed)
	return m.readErr("set speed")
}

// Reset sets the time to the current real time and resumes the normal flow of time.
func (m *Mocker) Reset(ctx context.Context) error {
	m.setOffset(ctx, 0)
	m.setSpeed(ctx, 1.0)
	return m.readErr("reset")
}
