package migrate

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/target/goalert/lock"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

//go:embed migrations
var fs embed.FS

func migrationIDs() []string {
	files, err := fs.ReadDir("migrations")
	if err != nil {
		panic(err)
	}

	ids := make([]string, 0, len(files))
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		ids = append(ids, strings.TrimPrefix(f.Name(), "migrations/"))
	}

	return ids
}

// Names will return all AssetNames without the timestamps and extensions
func Names() []string {
	uniq := make(map[string]struct{})
	var names []string

	for _, id := range migrationIDs() {
		name := migrationName(id)
		if _, ok := uniq[name]; ok {
			panic("duplicate migration name " + name)
		}
		uniq[name] = struct{}{}

		names = append(names, name)
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
	for i, id := range migrationIDs() {
		if migrationName(id) == name {
			return i, id
		}
	}
	return -1, ""
}

// VerifyAll will verify all migrations have already been applied.
func VerifyAll(ctx context.Context, url string) error {
	ids := migrationIDs()
	targetIndex := len(ids) - 1
	targetID := ids[targetIndex]
	targetName := migrationName(targetID)

	conn, err := getConn(ctx, url)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	var hasLatest bool
	err = conn.QueryRow(ctx, `select true from gorp_migrations where id = $1`, targetID).Scan(&hasLatest)
	if !errors.Is(err, pgx.ErrNoRows) && err != nil {
		return err
	}
	if hasLatest {
		return nil
	}

	return errors.Errorf("latest migration '%s' has not been applied", targetName)
}

// ApplyAll will atomically perform all UP migrations.
func ApplyAll(ctx context.Context, url string) (int, error) {
	return Up(ctx, url, "")
}

func getConn(ctx context.Context, url string) (*pgx.Conn, error) {
	var conn *pgx.Conn
	err := retry.DoTemporaryError(func(int) error {
		var err error
		conn, err = pgx.Connect(ctx, url)
		return err
	},
		retry.Limit(12),
		retry.FibBackoff(time.Millisecond*100),
	)
	if err != nil {
		return nil, err
	}

	_, err = conn.Exec(ctx, `set lock_timeout = 15000`)
	if err != nil {
		conn.Close(ctx)
		return nil, errors.Wrap(err, "set lock timeout")
	}

	return conn, nil
}

func aquireLock(ctx context.Context, conn *pgx.Conn) error {
	for {
		_, err := conn.Exec(ctx, `select pg_advisory_lock($1)`, lock.GlobalMigrate)
		if err == nil {
			return nil
		}
		// 55P03 is lock_not_available
		// https://www.postgresql.org/docs/9.6/static/errcodes-appendix.html
		//
		// If the lock gets a timeout, terminate stale backends and try again.
		if e := sqlutil.MapError(err); e != nil && e.Code == "55P03" {
			log.Log(ctx, errors.Wrap(err, "get migration lock (will retry)"))
			_, err = conn.Exec(ctx, `
				select pg_terminate_backend(l.pid)
				from pg_locks l
				join pg_stat_activity act on act.pid = l.pid and state = 'idle' and state_change < now() - '30 seconds'::interval
				where locktype = 'advisory' and objid = $1 and granted
			`, lock.GlobalMigrate)
			if err != nil {
				conn.Close(ctx)
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

	_, err = conn.Exec(ctx, `
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
func Up(ctx context.Context, url, targetName string) (int, error) {
	if targetName == "" {
		names := Names()
		targetName = names[len(names)-1]
	}
	targetIndex, targetID := migrationID(targetName)
	if targetIndex == -1 {
		return 0, errors.Errorf("unknown migration target name '%s'", targetName)
	}

	conn, err := getConn(ctx, url)
	if err != nil {
		return 0, err
	}
	defer conn.Close(ctx)

	var hasLatest bool
	err = ensureTableQuery(ctx, conn, func() error {
		return conn.QueryRow(ctx, `select true from gorp_migrations where id = $1`, targetID).Scan(&hasLatest)
	})
	if err == nil && hasLatest {
		return 0, nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
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

	rows, err := conn.Query(ctx, `select id from gorp_migrations order by id`)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}
	defer rows.Close()

	ids := migrationIDs()
	i := -1
	for rows.Next() {
		i++
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return 0, errors.Wrap(err, "scan applied migrations")
		}
		if ids[i] != id {
			return 0, errors.Errorf("migration mismatch db has '%s' but expected '%s'", id, ids[i])
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
	defer conn.Close(ctx)

	var latest string
	err = ensureTableQuery(ctx, conn, func() error {
		return conn.QueryRow(ctx, `select id from gorp_migrations order by id desc limit 1`).Scan(&latest)
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

	rows, err := conn.Query(ctx, `select id from gorp_migrations where id > $1 order by id desc`, targetID)
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

func readMigration(id string) ([]byte, error) {
	data, err := fs.ReadFile("migrations/" + id)
	if err != nil {
		return nil, fmt.Errorf("read 'migrations/%s': %w", id, err)
	}
	return data, nil
}

// DumpMigrations will attempt to write all migration files to the specified directory
func DumpMigrations(dest string) error {
	for _, id := range migrationIDs() {
		fullPath := filepath.Join(dest, "migrations", id)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		data, err := readMigration(id)
		if err != nil {
			return err
		}
		err = os.WriteFile(fullPath, data, 0o644)
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
	for _, id := range migrationIDs() {
		var m migration
		m.ID = id
		m.Name = migrationName(id)
		data, err := readMigration(id)
		if err != nil {
			return nil, err
		}

		var up, down strings.Builder
		var isUp, isDown bool

		r := bufio.NewScanner(bytes.NewReader(data))
		for r.Scan() {
			line := r.Text()
			if strings.HasPrefix(line, "-- +migrate Up") {
				isUp = true
				isDown = false
				m.Up.disableTx = strings.Contains(line, "notransaction")
				continue
			}
			if strings.HasPrefix(line, "-- +migrate Down") {
				isUp = false
				isDown = true
				m.Down.disableTx = strings.Contains(line, "notransaction")
				continue
			}
			switch {
			case isUp:
				up.WriteString(line)
				up.WriteString("\n")
			case isDown:
				down.WriteString(line)
				down.WriteString("\n")
			}
			// ignore other lines
		}

		m.Up.statements = sqlutil.SplitQuery(up.String())
		m.Up.isUp = true
		m.Up.migration = &m

		m.Down.statements = sqlutil.SplitQuery(down.String())
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
		_, err := c.Exec(ctx, stmt)
		if err != nil {
			return errors.Wrapf(err, "statement #%d\n%s", i+1, stmt)
		}
	}

	_, err := c.Exec(ctx, step.doneStmt(), step.ID)
	if err != nil {
		return errors.Wrap(err, "update gorp_migrations")
	}

	return nil
}

func (step migrationStep) apply(ctx context.Context, c *pgx.Conn) error {
	if step.disableTx {
		return step.applyNoTx(ctx, c)
	}

	tx, err := c.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "begin tx")
	}
	defer sqlutil.RollbackContext(ctx, "migrate: apply", tx)

	// tx applies to the connection, so NoTx
	// will execute correctly.
	err = step.applyNoTx(ctx, c)
	if err != nil {
		return err
	}

	return errors.Wrap(tx.Commit(ctx), "commit")
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
