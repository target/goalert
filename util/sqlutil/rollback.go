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
	err := tx.Rollback()
	switch {
	case err == nil:
	case errors.Is(err, sql.ErrTxDone):
	case errors.Is(err, sql.ErrConnDone):
	default:
		log.Log(ctx, fmt.Errorf("%s: tx rollback: %w", errMsg, err))
	}
}

// RollbackContext provides the same functionality as Rollback, but for a pgx.Tx.
func RollbackContext(ctx context.Context, errMsg string, tx pgx.Tx) {
	err := tx.Rollback(ctx)
	switch {
	case err == nil:
	case errors.Is(err, context.Canceled):
	case errors.Is(err, pgx.ErrTxClosed):
	default:
		log.Log(ctx, fmt.Errorf("%s: tx rollback: %w", errMsg, err))
	}
}
