package sqlutil

import (
	"context"

	"gorm.io/gorm"
)

type ctxKey string

const (
	ctxKeyDB = ctxKey("sqlutil")
)

func FromContext(ctx context.Context) *gorm.DB { return ctx.Value(ctxKeyDB).(*gorm.DB) }
func Context(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, ctxKeyDB, db)
}

func Transaction(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	db := FromContext(ctx)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(tx.Statement.Context, ctxKeyDB, tx), tx)
	})
}

// Debug will enable debug logging for the given context.
func Debug(ctx context.Context) context.Context { return Context(ctx, FromContext(ctx).Debug()) }
