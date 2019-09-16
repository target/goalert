package migrate

//go:generate go run ../devtools/inliner -pkg $GOPACKAGE ./migrations/*.sql

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/pkg/errors"
	"github.com/rubenv/sql-migrate/sqlparse"
	"github.com/target/goalert/lock"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

// Names will return all AssetNames without the timestamps and extensions
func Names() []string {
	uniq := make(map[string]struct{})
	var names []string
	// Strip off "migrations/timestamp" and ".sql" file extension
	for _, b := range Files {
		name := migrationName(b.Name)
		if _, ok := uniq[name]; ok {
			panic("duplicate migation name " + name)
		}
		uniq[name] = struct{}{}

		names = append(names, migrationName(b.Name))
	}
	return names
}

func migrationName(file string) string {
	file = strings.TrimPrefix(file, "migrations/")
	// trim the timestamp, including the trailing hyphen
	// Example : 20170808110638-user-email.sql would become user-email.sql
	file = file[15:]
	file = strings.TrimSuffix(file, ".sql")
	return file
}
func migrationID(name string) (int, string) {
	for i, b := range Files {
		if migrationName(b.Name) == name {
			return i, strings.TrimPrefix(b.Name, "migrations/")
		}
	}
	return -1, ""
}

// ApplyAll will atomically perform all UP migrations.
func ApplyAll(ctx context.Context, url string) (int, error) {
	return Up(ctx, url, "")
}

func getConn(ctx context.Context, url string) (*pgx.Conn, error) {
	cfg, err := pgx.ParseConnectionString(url)
	if err != nil {
		return nil, err
	}

	conn, err := pgx.Connect(cfg)
	if err != nil {
		return nil, err
	}

	_, err = conn.ExecEx(ctx, `set lock_timeout = 15000`, nil)
	if err != nil {
		conn.Close()
		return nil, errors.Wrap(err, "set lock timeout")
	}

	return conn, nil

}
func aquireLock(ctx context.Context, conn *pgx.Conn) error {
	for {
		_, err := conn.ExecEx(ctx, `select pg_advisory_lock($1)`, nil, lock.GlobalMigrate)
		if err == nil {
			return nil
		}
		// 55P03 is lock_not_available
		// https://www.postgresql.org/docs/9.6/static/errcodes-appendix.html
		//
		// If the lock gets a timeout, terminate stale backends and try again.
		if e := sqlutil.MapError(err); e != nil && e.Code == "55P03" {
			log.Log(ctx, errors.Wrap(err, "get migration lock (will retry)"))
			_, err = conn.ExecEx(ctx, `
				select pg_terminate_backend(l.pid)
				from pg_locks l
				join pg_stat_activity act on act.pid = l.pid and state = 'idle' and state_change < now() - '30 seconds'::interval
				where locktype = 'advisory' and objid = $1 and granted
			`, nil, lock.GlobalMigrate)
			if err != nil {
				conn.Close()
				return errors.Wrap(err, "terminate stale backends")
			}
			continue
		}
	}
}

func ensureTableQuery(ctx context.Context, conn *pgx.Conn, fn func() error) error {
	err := fn()
	if err == nil {
		return nil
	}
	// 42P01 is undefined_table
	// https://www.postgresql.org/docs/9.6/static/errcodes-appendix.html
	if e := sqlutil.MapError(err); e != nil && e.Code == "42P01" {
		// continue
	} else {
		return err
	}

	_, err = conn.ExecEx(ctx, `
		CREATE TABLE IF NOT EXISTS gorp_migrations (
			id text PRIMARY KEY,
			applied_at timestamp with time zone
		)
	`, nil)
	if err != nil {
		return err
	}
	return fn()
}

// Up will apply all migrations up to, and including, targetName.
// If targetName is empty, all available migrations are applied.
func Up(ctx context.Context, url, targetName string) (int, error) {
	if targetName == "" {
		targetName = migrationName(Files[len(Files)-1].Name)
	}
	targetIndex, targetID := migrationID(targetName)
	if targetIndex == -1 {
		return 0, errors.Errorf("unknown migration target name '%s'", targetName)
	}

	conn, err := getConn(ctx, url)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	var hasLatest bool
	err = ensureTableQuery(ctx, conn, func() error {
		return conn.QueryRowEx(ctx, `select true from gorp_migrations where id = $1`, &pgx.QueryExOptions{
			ParameterOIDs: []pgtype.OID{pgtype.TextOID},
		}, targetID).Scan(&hasLatest)
	})
	if err == nil && hasLatest {
		return 0, nil
	}
	if err != nil && err != pgx.ErrNoRows {
		return 0, err
	}

	migrations, err := parseMigrations()
	if err != nil {
		return 0, err
	}
	err = aquireLock(ctx, conn)
	if err != nil {
		return 0, err
	}

	rows, err := conn.QueryEx(ctx, `select id from gorp_migrations order by id`, nil)
	if err != nil && err != pgx.ErrNoRows {
		return 0, err
	}
	defer rows.Close()

	i := -1
	for rows.Next() {
		i++
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return 0, errors.Wrap(err, "scan applied migrations")
		}
		if strings.TrimPrefix(Files[i].Name, "migrations/") != id {
			return 0, errors.Errorf("migration mismatch db has '%s' but expected '%s'", id, strings.TrimPrefix(Files[i].Name, "migrations/"))
		}
	}

	return performMigrations(ctx, conn, true, migrations[i+1:targetIndex+1])
}

// Down will roll back all migrations up to, but NOT including, targetName.
//
// If the DB contains unknown migrations, err is returned.
func Down(ctx context.Context, url, targetName string) (int, error) {
	targetIndex, targetID := migrationID(targetName)
	if targetIndex == -1 {
		return 0, errors.Errorf("unknown migration target name '%s'", targetName)
	}

	conn, err := getConn(ctx, url)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	var latest string
	err = ensureTableQuery(ctx, conn, func() error {
		return conn.QueryRowEx(ctx, `select id from gorp_migrations order by id desc limit 1`, nil).Scan(&latest)
	})
	if err != nil {
		return 0, err
	}
	if latest == targetID {
		return 0, nil
	}

	migrations, err := parseMigrations()
	if err != nil {
		return 0, err
	}
	byID := make(map[string]migration)
	for _, m := range migrations {
		byID[m.ID] = m
	}

	err = aquireLock(ctx, conn)
	if err != nil {
		return 0, err
	}

	rows, err := conn.QueryEx(ctx, `select id from gorp_migrations where id > $1 order by id desc`, nil, targetID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	migrations = migrations[:0]
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return 0, err
		}
		m, ok := byID[id]
		if !ok {
			return 0, errors.Errorf("could not find db migration '%s' to roll back", id)
		}
		migrations = append(migrations, m)
	}

	return performMigrations(ctx, conn, false, migrations)
}

// DumpMigrations will attempt to write all migration files to the specified directory
func DumpMigrations(dest string) error {
	for _, file := range Files {
		fullPath := filepath.Join(dest, filepath.Base(file.Name))
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		err := ioutil.WriteFile(fullPath, file.Data(), 0644)
		if err != nil {
			return errors.Wrapf(err, "write to %s", fullPath)
		}
	}
	return nil
}

type migration struct {
	ID   string
	Name string

	Up   migrationStep
	Down migrationStep
}
type migrationStep struct {
	statements []string
	disableTx  bool
	isUp       bool
	*migration
}

func parseMigrations() ([]migration, error) {
	var migrations []migration
	for _, file := range Files {
		var m migration
		m.ID = strings.TrimPrefix(file.Name, "migrations/")
		m.Name = migrationName(file.Name)
		p, err := sqlparse.ParseMigration(bytes.NewReader(file.Data()))
		if err != nil {
			return nil, errors.Wrapf(err, "parse %s", m.ID)
		}
		m.Up.statements = p.UpStatements
		m.Up.disableTx = p.DisableTransactionUp
		m.Up.isUp = true
		m.Up.migration = &m

		m.Down.statements = p.DownStatements
		m.Down.disableTx = p.DisableTransactionDown
		m.Down.migration = &m

		migrations = append(migrations, m)
	}
	return migrations, nil
}

const (
	deleteMigrationRecord = `delete from gorp_migrations where id = $1`
	insertMigrationRecord = `insert into gorp_migrations (id, applied_at) values ($1, now())`
)

func (step migrationStep) doneStmt() string {
	if step.isUp {
		return insertMigrationRecord
	}
	return deleteMigrationRecord
}
func (step migrationStep) applyNoTx(ctx context.Context, c *pgx.Conn) error {
	for i, stmt := range step.statements {
		_, err := c.ExecEx(ctx, stmt, nil)
		if err != nil && err != pgx.ErrNoRows {
			return errors.Wrapf(err, "statement #%d", i)
		}
	}

	_, err := c.ExecEx(ctx, step.doneStmt(), &pgx.QueryExOptions{ParameterOIDs: []pgtype.OID{pgtype.TextOID}}, step.ID)
	if err != nil {
		return errors.Wrap(err, "update gorp_migrations")
	}

	return nil
}

func (step migrationStep) apply(ctx context.Context, c *pgx.Conn) (err error) {
	if step.disableTx {
		return step.applyNoTx(ctx, c)
	}

	tx, err := c.BeginEx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin tx")
	}
	defer tx.Rollback()

	b := tx.BeginBatch()
	defer b.Close()

	for _, stmt := range step.statements {
		b.Queue(stmt, nil, nil, nil)
	}
	b.Queue(step.doneStmt(), []interface{}{step.ID}, []pgtype.OID{pgtype.TextOID}, nil)

	err = b.Send(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "send statements")
	}
	err = b.Close()
	if err != nil && err != pgx.ErrNoRows {
		return errors.Wrap(err, "execute migration steps")
	}

	return tx.CommitEx(ctx)
}

func performMigrations(ctx context.Context, c *pgx.Conn, applyUp bool, migrations []migration) (int, error) {
	typ := "DOWN"
	if applyUp {
		typ = "UP"
	}

	for i, m := range migrations {
		step := m.Down
		if applyUp {
			step = m.Up
		}

		s := time.Now()
		err := step.apply(ctx, c)
		if err != nil {
			return i, errors.Wrapf(err, "apply '%s'", m.Name)
		}
		log.Debugf(ctx, "Applied %s migration '%s' in %s", typ, m.Name, time.Since(s).Truncate(time.Millisecond))
	}

	return len(migrations), nil
}
