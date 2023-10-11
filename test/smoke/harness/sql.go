package harness

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/target/goalert/util/sqlutil"
)

// ExecSQL will execute all queries one-by-one.
func ExecSQL(ctx context.Context, url string, query string) error {
	queries := sqlutil.SplitQuery(query)

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	for _, q := range queries {
		_, err := conn.Exec(ctx, q)
		if err != nil {
			return err
		}
	}

	return nil
}

// ExecSQLBatch will execute all queries in a transaction by sending them all at once.
func ExecSQLBatch(ctx context.Context, url string, query string) error {
	queries := sqlutil.SplitQuery(query)

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "set statement_timeout = 3000")
	if err != nil {
		return err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer sqlutil.RollbackContext(ctx, "harness: exec sql", tx)

	b := &pgx.Batch{}
	for _, q := range queries {
		b.Queue(q)
	}

	err = tx.SendBatch(ctx, b).Close()
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// SQLRollback will rollback the transaction for cleanup, failing the test on error.
func SQLRollback(t *testing.T, errMsg string, tx *sql.Tx) {
	err := tx.Rollback()
	switch {
	case err == nil:
	case errors.Is(err, sql.ErrTxDone):
	case errors.Is(err, sql.ErrConnDone):
	default:
		t.Fatalf("ERROR: %s: tx rollback: %v", errMsg, err)
	}
}
