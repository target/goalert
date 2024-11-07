package app

import (
	"context"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/riverqueue/river/rivertype"
	"github.com/target/goalert/config"
	"github.com/target/goalert/engine/cleanupmanager"
	"github.com/target/goalert/util/log"
)

type riverErrs struct{}

func (r *riverErrs) HandleError(ctx context.Context, job *rivertype.JobRow, err error) *river.ErrorHandlerResult {
	ctx = log.WithField(ctx, "job.queue", job.Queue)
	ctx = log.WithField(ctx, "job.id", job.ID)
	ctx = log.WithField(ctx, "job.kind", job.Kind)
	log.Log(ctx, err)

	return nil
}

func (r *riverErrs) HandlePanic(ctx context.Context, job *rivertype.JobRow, panicVal any, trace string) *river.ErrorHandlerResult {
	ctx = log.WithField(ctx, "job.queue", job.Queue)
	ctx = log.WithField(ctx, "job.id", job.ID)
	ctx = log.WithField(ctx, "job.kind", job.Kind)
	ctx = log.WithField(ctx, "trace", trace)
	log.Log(ctx, fmt.Errorf("panic: %v", panicVal))

	return nil
}

func (app *App) initRiver(ctx context.Context) error {
	w := river.NewWorkers()

	err := cleanupmanager.AddWorkers(ctx, app.db, w)
	if err != nil {
		return err
	}

	app.River, err = river.NewClient(riverdatabasesql.New(app.db), &river.Config{
		Logger:  log.NewSlog(app.cfg.Logger),
		Workers: w,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		ErrorHandler: &riverErrs{},
	})
	if err != nil {
		return err
	}

	cfg := config.FromContext(ctx)
	err = cleanupmanager.InitRiverClient(cfg, app.db, app.River)
	if err != nil {
		return err
	}

	return nil
}
