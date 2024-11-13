package app

import (
	"context"
	"log/slog"

	"github.com/riverqueue/river"
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
		Logger:  app.Logger,
		Workers: app.RiverWorkers,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		ErrorHandler: &riverErrs{
			Logger: app.Logger,
		},
	})
	if err != nil {
		return err
	}

	opts := &riverui.ServerOpts{
		Prefix: "/admin/riverui",
		DB:     app.pgx,
		Client: app.River,
		Logger: app.Logger,
	}
	app.RiverUI, err = riverui.NewServer(opts)
	if err != nil {
		return err
	}

	return nil
}
