package processinglock

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/lock"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

// A Lock is used to start "locked" transactions.
type Lock struct {
	cfg Config
	db  *sql.DB
}
type txBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// NewLock will return a new Lock for the given Config.
func NewLock(ctx context.Context, db *sql.DB, cfg Config) (*Lock, error) {
	vers, err := gadb.New(db).ProcReadModuleVersion(ctx, gadb.EngineProcessingType(cfg.Type))
	if err != nil {
		return nil, fmt.Errorf("read module version: %w", err)
	}

	if vers != int32(cfg.Version) {
		log.Log(ctx, fmt.Errorf("engine module disabled: %s: version mismatch: expected=%d got=%d", cfg.Type, cfg.Version, vers))

		// Log the error, but continue operation.
		//
		// It is valid for individual engine modules to be disabled, but we
		// should not block the rest of the system, so we don't return here.
		//
		// However, we do log it as an error as it _will_ cause the lock to
		// fail to be acquired. This means if all instances of the engine have
		// the module disabled (like scheduling or messaging) then no schedules
		// will update or messages will be sent, respectively.
		//
		// We only do this at startup, rather than on every lock acquisition,
		// to avoid spamming the logs with the same error. Also because during
		// an upgrade, existing engine instances are expected to be running, and
		// thus stop processing new work for the now incompatible module.
		//
		// Starting an _old_ engine instance with a new module version is not
		// as common, and we want to inform the operator that this is happening,
		// as it is likely part of a rollback or other unexpected situation.
		//
		// In those cases, the operator should be aware that the engine is not
		// processing work as expected, because the database is still using a
		// structure that the older code does not understand.
	}

	return &Lock{
		db:  db,
		cfg: cfg,
	}, nil
}

func (l *Lock) _BeginTx(ctx context.Context, b txBeginner, opts *sql.TxOptions, shared bool) (*sql.Tx, error) {
	tx, err := b.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	q := gadb.New(tx)

	if shared {
		// Shared transactions are run by retryable background jobs, so rather
		// than block indefinitely behind an exclusive holder of the module row
		// lock (e.g. a legacy UpdateAll using FOR UPDATE NOWAIT), bound the
		// wait so the job fails fast and can be retried later.
		//
		// This must be set before ProcAcquireModuleSharedLock (FOR SHARE),
		// which is the only statement in this function that can wait on a
		// lock. Setting it here, rather than after _BeginTx returns, is the
		// whole point: lock_timeout only applies to statements issued after
		// it, and the wait we care about is the FOR SHARE below.
		//
		// The advisory lock acquired first uses pg_try_advisory_xact_lock_shared,
		// which is a try-lock and never waits, so lock_timeout has no effect on
		// it (it still returns false immediately if the lock is unavailable).
		//
		// When the timeout expires Postgres raises SQLSTATE 55P03
		// (lock_not_available), the same code as NOWAIT, and it is mapped to
		// ErrNoLock below.
		_, err = tx.ExecContext(ctx, `SET LOCAL lock_timeout = 8000`)
		if err != nil {
			sqlutil.Rollback(ctx, "processing lock: begin", tx)
			return nil, err
		}
	}

	// Ensure the engine isn't running or that it waits for migrations to complete.
	gotAdvLock, err := q.ProcSharedAdvisoryLock(ctx, int64(lock.GlobalMigrate))
	if err != nil {
		sqlutil.Rollback(ctx, "processing lock: begin", tx)
		return nil, err
	}
	if !gotAdvLock {
		sqlutil.Rollback(ctx, "processing lock: begin", tx)
		return nil, ErrNoLock
	}

	if shared {
		_, err = q.ProcAcquireModuleSharedLock(ctx, gadb.ProcAcquireModuleSharedLockParams{
			TypeID:  gadb.EngineProcessingType(l.cfg.Type),
			Version: l.cfg.Version,
		})
	} else {
		_, err = q.ProcAcquireModuleLockNoWait(ctx, gadb.ProcAcquireModuleLockNoWaitParams{
			TypeID:  gadb.EngineProcessingType(l.cfg.Type),
			Version: l.cfg.Version,
		})
	}
	if errors.Is(err, sql.ErrNoRows) {
		// wrong module version
		sqlutil.Rollback(ctx, "processing lock: begin", tx)
		return nil, ErrNoLock
	}
	if err != nil {
		sqlutil.Rollback(ctx, "processing lock: begin", tx)
		// 55P03 is lock_not_available, raised either by the `nowait` in the
		// exclusive query, or by lock_timeout expiring in the shared path.
		//
		// https://www.postgresql.org/docs/9.4/static/errcodes-appendix.html
		if sqlErr := sqlutil.MapError(err); sqlErr != nil && sqlErr.Code == "55P03" {
			return nil, ErrNoLock
		}
		return nil, err
	}

	return tx, nil
}

// WithTx will run the given function in a locked transaction, returning ErrNoLock immediately if the lock cannot be acquired.
func (l *Lock) WithTx(ctx context.Context, fn func(context.Context, *sql.Tx) error) error {
	tx, err := l._BeginTx(ctx, l.db, nil, false)
	if err != nil {
		return err
	}
	defer sqlutil.Rollback(ctx, "processing lock: with tx", tx)

	err = fn(ctx, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// WithTxShared will run the given function in a locked transaction that is compatible with other shared locks.
//
// Acquiring the module lock blocks for at most lock_timeout (8s); if an
// exclusive holder (e.g. a legacy UpdateAll) still has the row after that,
// ErrNoLock is returned without calling fn. Callers can detect this with
// IsNoLock. The same lock_timeout remains in effect for statements run by fn.
//
// Note: this is a behavior change for existing River workers using this
// method (rotationmanager, statusmgr, cleanupmanager, signalmgr). Previously
// the module lock was acquired before lock_timeout was set, so a blocked
// worker waited until River's job timeout (1 minute by default), then failed
// and burned an attempt. They now fail fast with ErrNoLock after 8s; River
// still logs the error and retries the job on its normal schedule.
func (l *Lock) WithTxShared(ctx context.Context, fn func(context.Context, *sql.Tx) error) error {
	tx, err := l._BeginTx(ctx, l.db, nil, true)
	if err != nil {
		return err
	}
	defer sqlutil.Rollback(ctx, "processing lock: with tx (shared)", tx)

	err = fn(ctx, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (l *Lock) _Exec(ctx context.Context, b txBeginner, stmt *sql.Stmt, args ...interface{}) (sql.Result, error) {
	tx, err := l._BeginTx(ctx, b, nil, false)
	if err != nil {
		return nil, err
	}
	defer sqlutil.Rollback(ctx, "processing lock: exec", tx)

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
	return l._BeginTx(ctx, l.db, opts, false)
}

// Exec will run ExecContext on the statement, wrapped in a locked transaction.
func (l *Lock) Exec(ctx context.Context, stmt *sql.Stmt, args ...interface{}) (sql.Result, error) {
	return l._Exec(ctx, l.db, stmt, args...)
}
