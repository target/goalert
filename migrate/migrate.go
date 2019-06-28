package migrate

//go:generate go run ../devtools/inliner -pkg $GOPACKAGE ./migrations/*.sql

import (
	"bytes"
	"context"
	"database/sql"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/rubenv/sql-migrate/sqlparse"
	"github.com/target/goalert/lock"
	"github.com/target/goalert/util/log"
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
func ApplyAll(ctx context.Context, db *sql.DB) (int, error) {
	return Up(ctx, db, "")
}

func getConn(ctx context.Context, db *sql.DB) (*sql.Conn, error) {
	c, err := db.Conn(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get db conn")
	}

	_, err = c.ExecContext(ctx, `set lock_timeout = 15000`)
	if err != nil {
		releaseConn(c)
		return nil, errors.Wrap(err, "set lock timeout")
	}

	for {
		_, err = c.ExecContext(ctx, `select pg_advisory_lock($1)`, lock.GlobalMigrate)
		if err == nil {
			return c, nil
		}
		// 55P03 is lock_not_available
		// https://www.postgresql.org/docs/9.6/static/errcodes-appendix.html
		//
		// If the lock gets a timeout, terminate stale backends and try again.
		if pErr, ok := err.(*pq.Error); ok && pErr.Code == "55P03" {
			log.Log(ctx, errors.Wrap(err, "get migration lock (will retry)"))
			_, err = c.ExecContext(ctx, `
				select pg_terminate_backend(l.pid)
				from pg_locks l
				join pg_stat_activity act on act.pid = l.pid and state = 'idle' and state_change < now() - '30 seconds'::interval
				where locktype = 'advisory' and objid = $1 and granted
			`, lock.GlobalMigrate)
			if err != nil {
				releaseConn(c)
				return nil, errors.Wrap(err, "terminate stale backends")
			}
			continue
		}

		releaseConn(c)
		return nil, errors.Wrap(err, "get migration lock")
	}

}
func releaseConn(c *sql.Conn) {
	c.ExecContext(context.Background(), `select pg_advisory_unlock($1)`, lock.GlobalMigrate)
	c.ExecContext(context.Background(), `set lock_timeout to default`)
	c.Close()
}

func ensureTableQuery(ctx context.Context, db *sql.DB, fn func() error) error {
	err := fn()
	if err == nil {
		return nil
	}
	// 42P01 is undefined_table
	// https://www.postgresql.org/docs/9.6/static/errcodes-appendix.html
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "42P01" {
		// continue
	} else if pgxErr, ok := err.(pgx.PgError); ok && pgxErr.Code == "42P01" {
		// continue
	} else {
		return err
	}

	c, err := getConn(ctx, db)
	if err != nil {
		return err
	}
	defer releaseConn(c)
	_, err = c.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS gorp_migrations (
			id text PRIMARY KEY,
			applied_at timestamp with time zone
		)
	`)
	if err != nil {
		return err
	}
	return fn()
}

// Up will apply all migrations up to, and including, targetName.
// If targetName is empty, all available migrations are applied.
func Up(ctx context.Context, db *sql.DB, targetName string) (int, error) {
	if targetName == "" {
		targetName = migrationName(Files[len(Files)-1].Name)
	}
	targetIndex, targetID := migrationID(targetName)
	if targetIndex == -1 {
		return 0, errors.Errorf("unknown migration target name '%s'", targetName)
	}

	var hasLatest bool
	err := ensureTableQuery(ctx, db, func() error {
		return db.QueryRowContext(ctx, `select true from gorp_migrations where id = $1`, targetID).Scan(&hasLatest)
	})
	if err == nil && hasLatest {
		return 0, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	migrations, err := parseMigrations()
	if err != nil {
		return 0, err
	}

	c, err := getConn(ctx, db)
	if err != nil {
		return 0, err
	}
	defer releaseConn(c)

	rows, err := c.QueryContext(ctx, `select id from gorp_migrations order by id`)
	if err != nil && err != sql.ErrNoRows {
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

	return performMigrations(ctx, c, true, migrations[i+1:targetIndex+1])
}

// Down will roll back all migrations up to, but NOT including, targetName.
//
// If the DB contains unknown migrations, err is returned.
func Down(ctx context.Context, db *sql.DB, targetName string) (int, error) {
	targetIndex, targetID := migrationID(targetName)
	if targetIndex == -1 {
		return 0, errors.Errorf("unknown migration target name '%s'", targetName)
	}

	var latest string
	err := ensureTableQuery(ctx, db, func() error {
		return db.QueryRowContext(ctx, `select id from gorp_migrations order by id desc limit 1`).Scan(&latest)
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

	c, err := getConn(ctx, db)
	if err != nil {
		return 0, err
	}
	defer releaseConn(c)
	rows, err := c.QueryContext(ctx, `select id from gorp_migrations where id > $1 order by id desc`, targetID)
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

	return performMigrations(ctx, c, false, migrations)
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
	*sqlparse.ParsedMigration
}

func parseMigrations() ([]migration, error) {
	var migrations []migration
	var err error
	for _, file := range Files {
		var m migration
		m.ID = strings.TrimPrefix(file.Name, "migrations/")
		m.Name = migrationName(file.Name)
		m.ParsedMigration, err = sqlparse.ParseMigration(bytes.NewReader(file.Data()))
		if err != nil {
			return nil, errors.Wrapf(err, "parse %s", m.ID)
		}

		migrations = append(migrations, m)
	}
	return migrations, nil
}

func (m migration) apply(ctx context.Context, c *sql.Conn, up bool) (err error) {
	var tx *sql.Tx
	type execer interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}
	s := time.Now()
	typ := "UP"
	if !up {
		typ = "DOWN"
	}
	ex := execer(c)
	if up && !m.DisableTransactionUp || !up && !m.DisableTransactionDown {
		tx, err = c.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		ex = tx
	}

	stmts := m.UpStatements
	if !up {
		stmts = m.DownStatements
	}
	for _, s := range stmts {
		_, err = ex.ExecContext(ctx, s)
		if err != nil {
			return err
		}
	}

	if up {
		_, err = ex.ExecContext(ctx, `insert into gorp_migrations (id, applied_at) values ($1, now())`, m.ID)
	} else {
		_, err = ex.ExecContext(ctx, `delete from gorp_migrations where id = $1`, m.ID)
	}
	if err != nil {
		return err
	}

	if tx != nil {
		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	log.Debugf(ctx, "Applied %s migration '%s' in %s", typ, m.Name, time.Since(s).Truncate(time.Millisecond))

	return nil
}
func performMigrations(ctx context.Context, c *sql.Conn, applyUp bool, migrations []migration) (int, error) {
	for i, m := range migrations {
		err := m.apply(ctx, c, applyUp)
		if err != nil {
			return i, err
		}
	}
	return len(migrations), nil
}
