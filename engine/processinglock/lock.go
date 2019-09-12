package processinglock

import (
	"context"
	"database/sql"

	"github.com/target/goalert/lock"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"

	"go.opencensus.io/trace"
)

// A Lock is used to start "locked" transactions.
type Lock struct {
	cfg      Config
	db       *sql.DB
	lockStmt *sql.Stmt

	advLockStmt *sql.Stmt
}
type txBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// NewLock will return a new Lock for the given Config.
func NewLock(ctx context.Context, db *sql.DB, cfg Config) (*Lock, error) {
	p := &util.Prepare{Ctx: ctx, DB: db}
	return &Lock{
		db:          db,
		cfg:         cfg,
		advLockStmt: p.P(`select pg_try_advisory_xact_lock_shared($1)`),
		lockStmt: p.P(`
			select version
			from engine_processing_versions
			where type_id = $1
			for update nowait
		`),
	}, p.Err
}

func (l *Lock) _BeginTx(ctx context.Context, b txBeginner, opts *sql.TxOptions) (*sql.Tx, error) {
	ctx, sp := trace.StartSpan(ctx, "ProcessingLock.BeginTx")
	defer sp.End()
	l.cfg.decorateSpan(sp)

	tx, err := b.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Ensure the engine isn't running or that it waits for migrations to complete.
	var gotAdvLock bool
	err = tx.StmtContext(ctx, l.advLockStmt).QueryRowContext(ctx, lock.GlobalMigrate).Scan(&gotAdvLock)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if !gotAdvLock {
		tx.Rollback()
		return nil, ErrNoLock
	}

	var dbVersion int
	err = tx.StmtContext(ctx, l.lockStmt).QueryRowContext(ctx, l.cfg.Type).Scan(&dbVersion)
	if err != nil {
		tx.Rollback()
		// 55P03 is lock_not_available (due to the `nowait` in the query)
		//
		// https://www.postgresql.org/docs/9.4/static/errcodes-appendix.html
		if sqlErr := sqlutil.MapError(err); sqlErr != nil && sqlErr.Code == "55P03" {
			return nil, ErrNoLock
		}
		return nil, err
	}
	if dbVersion != l.cfg.Version {
		tx.Rollback()
		return nil, ErrNoLock
	}

	return tx, nil
}
func (l *Lock) _Exec(ctx context.Context, b txBeginner, stmt *sql.Stmt, args ...interface{}) (sql.Result, error) {
	tx, err := l._BeginTx(ctx, b, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	res, err := tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return res, nil
}

// BeginTx will start a transaction with the appropriate lock in place (based on Config).
func (l *Lock) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return l._BeginTx(ctx, l.db, opts)
}

// Exec will run ExecContext on the statement, wrapped in a locked transaction.
func (l *Lock) Exec(ctx context.Context, stmt *sql.Stmt, args ...interface{}) (sql.Result, error) {
	return l._Exec(ctx, l.db, stmt, args...)
}
