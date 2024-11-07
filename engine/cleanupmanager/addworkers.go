package cleanupmanager

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/riverqueue/river"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/util/log"
)

const (
	QueueName           = "engine-cleanup"
	SchedCleanupJobName = "engine-cleanup-schedules"
)

type ScheduleCleanupArgs struct {
	MaxAgeDays int
}

func (ScheduleCleanupArgs) Kind() string { return SchedCleanupJobName }

type ScheduleCleanupWorker struct {
	river.WorkerDefaults[ScheduleCleanupArgs]

	Logger *slog.Logger
	db     *gadb.Queries
}

func (w *ScheduleCleanupWorker) Work(ctx context.Context, job *river.Job[ScheduleCleanupArgs]) error {
	now, err := w.db.Now(ctx)
	if err != nil {
		return fmt.Errorf("get current DB time: %w", err)
	}

	cutoff := now.AddDate(0, 0, -job.Args.MaxAgeDays)

	var madeChanges bool

	rows, err := w.db.CleanMgrTrimOverrides(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("trim overrides: %w", err)
	}
	madeChanges = madeChanges || rows > 0
	w.Logger.DebugContext(ctx, "trimmed overrides", slog.Int64("rows", rows))

	rows, err = w.db.CleanMgrTrimSchedOnCall(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("trim schedule on-call: %w", err)
	}
	madeChanges = madeChanges || rows > 0
	w.Logger.DebugContext(ctx, "trimmed schedule on-call", slog.Int64("rows", rows))

	rows, err = w.db.CleanMgrTrimEPOnCall(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("trim escalation policy on-call: %w", err)
	}
	madeChanges = madeChanges || rows > 0
	w.Logger.DebugContext(ctx, "trimmed escalation policy on-call", slog.Int64("rows", rows))

	if madeChanges { // re-enqueue if we trimmed any overrides
		c := river.ClientFromContext[*sql.Tx](ctx)
		_, err := c.Insert(ctx, job.Args, &river.InsertOpts{
			Queue:      QueueName,
			UniqueOpts: river.UniqueOpts{ByQueue: true},
		})
		if err != nil {
			return fmt.Errorf("re-enqueue: %w", err)
		}
	}

	return nil
}

func AddWorkers(ctx context.Context, db gadb.DBTX, w *river.Workers) error {
	return river.AddWorkerSafely(w, &ScheduleCleanupWorker{db: gadb.New(db), Logger: log.NewSlog(log.FromContext(ctx))})
}

func InitRiverClient[T any](cfg config.Config, db gadb.DBTX, c *river.Client[T]) error {
	err := c.Queues().Add(QueueName, river.QueueConfig{MaxWorkers: 1})
	if err != nil {
		return err
	}
	cfg.Maintenance.ScheduleCleanupDays = 30

	if cfg.Maintenance.ScheduleCleanupDays > 0 {
		c.PeriodicJobs().Add(river.NewPeriodicJob(
			river.PeriodicInterval(time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return ScheduleCleanupArgs{MaxAgeDays: cfg.Maintenance.ScheduleCleanupDays}, &river.InsertOpts{
					Queue:      QueueName,
					UniqueOpts: river.UniqueOpts{ByQueue: true},
				}
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		))
	}

	return nil
}
