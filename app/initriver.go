package app

import (
	"context"
	"log/slog"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"
	"riverqueue.com/riverui"
)

type riverErrs struct {
	Logger *slog.Logger
}

func (r *riverErrs) HandleError(ctx context.Context, job *rivertype.JobRow, err error) *river.ErrorHandlerResult {
	r.Logger.ErrorContext(ctx, "Job returned error.",
		"job.queue", job.Queue,
		"job.id", job.ID,
		"job.kind", job.Kind,
		"err", err,
	)

	return nil
}

func (r *riverErrs) HandlePanic(ctx context.Context, job *rivertype.JobRow, panicVal any, trace string) *river.ErrorHandlerResult {
	r.Logger.ErrorContext(ctx, "Job panicked.",
		"job.queue", job.Queue,
		"job.id", job.ID,
		"job.kind", job.Kind,
		"panic", panicVal,
		"trace", trace,
	)

	return nil
}

type noopWorker struct{}

func (noopWorker) Kind() string { return "noop" }

// ignoreCancel is a slog.Handler that ignores log records with an "error" attribute of "context canceled".
type ignoreCancel struct{ h slog.Handler }

// Enabled implements the slog.Handler interface.
func (i *ignoreCancel) Enabled(ctx context.Context, level slog.Level) bool {
	return i.h.Enabled(ctx, level)
}

// Handle implements the slog.Handler interface.
func (i *ignoreCancel) Handle(ctx context.Context, rec slog.Record) error {
	var shouldIgnore bool
	rec.Attrs(func(a slog.Attr) bool {
		if a.Key == "error" && a.Value.String() == "context canceled" {
			shouldIgnore = true
		}
		return true
	})
	if shouldIgnore {
		return nil
	}
	return i.h.Handle(ctx, rec)
}

// WithContext implements the slog.Handler interface.
func (i *ignoreCancel) WithGroup(name string) slog.Handler {
	return &ignoreCancel{h: i.h.WithGroup(name)}
}

// WithAttrs implements the slog.Handler interface.
func (i *ignoreCancel) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ignoreCancel{h: i.h.WithAttrs(attrs)}
}

func (app *App) initRiver(ctx context.Context) error {
	app.RiverWorkers = river.NewWorkers()

	// TODO: remove once a worker is added that's not behind a feature flag
	//
	// Without this, it will complain about no workers being registered.
	river.AddWorker(app.RiverWorkers, river.WorkFunc(func(ctx context.Context, j *river.Job[noopWorker]) error {
		// Do something with the job
		return nil
	}))

	var err error
	app.River, err = river.NewClient(riverpgxv5.New(app.pgx), &river.Config{
		// River tends to log "context canceled" errors while shutting down
		Logger:  slog.New(&ignoreCancel{h: app.Logger.With("module", "river").Handler()}),
		Workers: app.RiverWorkers,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		ErrorHandler: &riverErrs{
			// The error handler logger is used differently than the main logger, so it should be separate, and doesn't need the wrapper.
			Logger: app.Logger.With("module", "river"),
		},
	})
	if err != nil {
		return err
	}

	app.RiverDBSQL, err = river.NewClient(riverdatabasesql.New(app.db), &river.Config{
		Logger:   slog.New(app.Logger.With("module", "river_dbsql").Handler()),
		PollOnly: true, // don't consume a connection trying to poll, since this client has no workers
	})
	if err != nil {
		return err
	}

	opts := &riverui.ServerOpts{
		Prefix: "/admin/riverui",
		DB:     app.pgx,
		Client: app.River,
		Logger: app.Logger.With("module", "riverui"),
	}
	app.RiverUI, err = riverui.NewServer(opts)
	if err != nil {
		return err
	}

	return nil
}
