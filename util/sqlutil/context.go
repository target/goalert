package sqlutil

import (
	"context"

	"gorm.io/gorm"
)

type ctxKey string

const (
	ctxKeyDB = ctxKey("sqlutil")
)

// FromContext will return the DB object from the given context.
func FromContext(ctx context.Context) *gorm.DB { return ctx.Value(ctxKeyDB).(*gorm.DB) }

// Context will return a new context with the given DB object.
func Context(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, ctxKeyDB, db)
}

// Transaction will run the given function in a transaction handling commit/rollback and sub-transactions.
func Transaction(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	db := FromContext(ctx)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(tx.Statement.Context, ctxKeyDB, tx), tx)
	})
}

// Debug will enable debug logging for the given context.
func Debug(ctx context.Context) context.Context { return Context(ctx, FromContext(ctx).Debug()) }
