package sqlutil

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/util/log"
)

// Rollback perform a DB rollback function
func Rollback(ctx context.Context, errMsg string, tx *sql.Tx) {
	if err := tx.Rollback(); err != nil {
		log.Log(ctx, fmt.Errorf("tx rollback issue at %s: %v", errMsg, err))
	}
}

// ContextRollback perform a DB rollback function with the context
func ContextRollback(ctx context.Context, errMsg string, tx pgx.Tx) {
	if err := tx.Rollback(ctx); err != nil {
		log.Log(ctx, fmt.Errorf("tx rollback issue at %s: %v", errMsg, err))
	}
}
