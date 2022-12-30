package sqlutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/util/log"
)

// Rollback perform a DB rollback function
func Rollback(ctx context.Context, errMsg string, tx *sql.Tx) {
	if err := tx.Rollback(); err != nil {
		if err != sql.ErrTxDone && err != sql.ErrConnDone {
			log.Log(ctx, fmt.Errorf("tx rollback issue at %s: %v", errMsg, err))
		}
	}
}

// RollbackTest will perform a DB rollback for use only within tests
func RollbackTest(t *testing.T, errMsg string, tx *sql.Tx) {
	if err := tx.Rollback(); err != nil {
		if err != sql.ErrTxDone && err != sql.ErrConnDone {
			t.Errorf("tx rollback issue at %s: %v", errMsg, err)
		}
	}
}

// ContextRollback perform a DB rollback function with the context
func ContextRollback(ctx context.Context, errMsg string, tx pgx.Tx) {
	if err := tx.Rollback(ctx); err != nil {
		if err != sql.ErrTxDone && err != sql.ErrConnDone {
			log.Log(ctx, fmt.Errorf("tx rollback issue at %s: %v", errMsg, err))
		}
	}
}
