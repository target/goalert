package event

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type Bus struct {
	l          *slog.Logger
	river      *river.Client[pgx.Tx]
	riverDBSQL *river.Client[*sql.Tx]

	b []any
}

func NewBus(l *slog.Logger) *Bus {
	return &Bus{l: l.With("component", "EventBus")}
}

func (b *Bus) SetRiver(r *river.Client[pgx.Tx]) { b.river = r }

func (b *Bus) SetRiverDBSQL(r *river.Client[*sql.Tx]) { b.riverDBSQL = r }

type (
	subBus[T, TTx comparable] struct {
		onBatch   []func(ctx context.Context, data []T) error
		onBatchTx []func(ctx context.Context, tx TTx, data []T) error
		l         *slog.Logger
	}
	nilTx any
)

func findBus[T, TTx comparable](b *Bus) *subBus[T, TTx] {
	for _, b := range b.b {
		s, ok := b.(*subBus[T, TTx])
		if ok {
			return s
		}
	}
	sub := subBus[T, TTx]{l: b.l}
	b.b = append(b.b, &sub)

	return &sub
}

func (s *subBus[T, TTx]) send(ctx context.Context, data []T) error {
	if len(s.onBatch) == 0 {
		var dataType T
		s.l.WarnContext(ctx, "no handlers for event",
			slog.Group("event",
				slog.String("dataType", fmt.Sprintf("%T", dataType))))
		return nil
	}
	var errs []error
	for _, fn := range s.onBatch {
		err := fn(ctx, data)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (s *subBus[T, TTx]) sendTx(ctx context.Context, tx TTx, data []T) error {
	if len(s.onBatchTx) == 0 {
		var dataType T
		s.l.WarnContext(ctx, "no handlers for transactional event",
			slog.Group("event",
				slog.String("txType", fmt.Sprintf("%T", tx)),
				slog.String("dataType", fmt.Sprintf("%T", dataType))))
		return nil
	}
	var errs []error
	for _, fn := range s.onBatchTx {
		err := fn(ctx, tx, data)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Send sends an event to the bus.
func Send[T comparable](ctx context.Context, b *Bus, event T) { SendMany(ctx, b, []T{event}) }

// Send sends an event to the bus.
func SendMany[T comparable](ctx context.Context, b *Bus, events []T) {
	SendManyTx[T, nilTx](ctx, b, nil, events)
}

// SendTx sends an event to the bus that is contingent on a transaction.
func SendTx[T, TTx comparable](ctx context.Context, b *Bus, tx TTx, event T) {
	SendManyTx(ctx, b, tx, []T{event})
}

// SendTx sends an event to the bus that is contingent on a transaction.
func SendManyTx[T, TTx comparable](ctx context.Context, b *Bus, tx TTx, events []T) {
	for _, e := range events {
		b.l.DebugContext(ctx, "send event", "event", e)
	}
	isNilTx := (any)(tx) == nil
	var err error
	sub := findBus[T, TTx](b)
	if isNilTx {
		err = sub.send(ctx, events)
	} else {
		err = sub.sendTx(ctx, tx, events)
	}

	if err != nil {
		b.l.ErrorContext(ctx, "send event failed", "error", err)
	}
}

type Event[T any] struct {
	Value *T

	// Tx is the transaction that the event is contingent on, if any.
	Tx pgx.Tx

	// SQLTx is the transaction that the event is contingent on, if any.
	SQLTx *sql.Tx
}

func OnEachBatch[T comparable](b *Bus, fn func(ctx context.Context, data []T) error) {
	sub := findBus[T, nilTx](b)
	sub.onBatch = append(sub.onBatch, fn)
}

func OnEachBatchTx[T, TTx comparable](b *Bus, fn func(ctx context.Context, tx TTx, data []T) error) {
	sub := findBus[T, TTx](b)
	sub.onBatchTx = append(sub.onBatchTx, fn)
}

func insertJobsTx[T, TTx any](ctx context.Context, rv *river.Client[TTx], tx TTx, data []T, newJobFn func(data T) (river.JobArgs, *river.InsertOpts)) error {
	args := make([]river.InsertManyParams, len(data))
	for i, d := range data {
		a, o := newJobFn(d)
		args[i] = river.InsertManyParams{
			Args:       a,
			InsertOpts: o,
		}
	}

	var err error
	var res []*rivertype.JobInsertResult
	isNilTx := (any)(tx) == nil
	if isNilTx {
		res, err = rv.InsertMany(ctx, args)
	} else {
		res, err = rv.InsertManyTx(ctx, tx, args)
	}
	if err != nil {
		return fmt.Errorf("insert many: %w", err)
	}

	for _, r := range res {
		if !r.UniqueSkippedAsDuplicate {
			continue
		}

		if isNilTx {
			_, err = rv.JobRetry(ctx, r.Job.ID)
		} else {
			_, err = rv.JobRetryTx(ctx, tx, r.Job.ID)
		}
		if err != nil {
			return fmt.Errorf("retry job: %w", err)
		}
	}
	return nil
}

func RegisterJobSource[T comparable](b *Bus, newJobFn func(data T) (river.JobArgs, *river.InsertOpts)) {
	OnEachBatch(b, func(ctx context.Context, data []T) error {
		return insertJobsTx(ctx, b.river, nil, data, newJobFn)
	})
	OnEachBatchTx(b, func(ctx context.Context, tx pgx.Tx, data []T) error {
		return insertJobsTx(ctx, b.river, tx, data, newJobFn)
	})
	OnEachBatchTx(b, func(ctx context.Context, tx *sql.Tx, data []T) error {
		return insertJobsTx(ctx, b.riverDBSQL, tx, data, newJobFn)
	})
}
