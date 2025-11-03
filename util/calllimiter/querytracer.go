package calllimiter

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type queryLimiterTracer struct{}

var _ pgx.QueryTracer = (*queryLimiterTracer)(nil)

var ErrQueryLimitReached = errors.New("query limit reached")

func (t *queryLimiterTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	l := FromContext(ctx)
	if l.Allow() {
		return ctx
	}

	ctx, cancel := context.WithCancelCause(ctx)
	cancel(&ErrCallLimitReached{NumCalls: l.NumCalls()})
	return ctx
}

func (t *queryLimiterTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
}

// SetConfigQueryLimiterSupport configures the pgxpool.Config to support query limiting.
// It sets the ConnConfig.Tracer to a QueryLimiter that will limit the number of queries
// executed concurrently and the total number of queries executed.
// This is useful for preventing excessive load on the database.
func SetConfigQueryLimiterSupport(cfg *pgxpool.Config) {
	cfg.ConnConfig.Tracer = &queryLimiterTracer{}
}
