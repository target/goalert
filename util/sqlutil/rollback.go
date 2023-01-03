package sqlutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/util/log"
)

// Rollback will roll back the transaction, logging any potential errors other than ErrTxDone and ErrConnDone,
// which are expected.
//
// Primarily, it's intended to be used with defer as an alternative to calling defer tx.Rollback() which
// ignores ALL errors.
func Rollback(ctx context.Context, errMsg string, tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) && !errors.Is(err, sql.ErrConnDone) {
		log.Log(ctx, fmt.Errorf("%s: tx rollback: %w", errMsg, err))
	}
}

// RollbackContext provides the same functionality as Rollback, but uses the pgx library rather than the standard
// sql library
func RollbackContext(ctx context.Context, errMsg string, tx pgx.Tx) {
	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, sql.ErrTxDone) && !errors.Is(err, sql.ErrConnDone) {
		log.Log(ctx, fmt.Errorf("%s: tx rollback: %w", errMsg, err))
	}
}
