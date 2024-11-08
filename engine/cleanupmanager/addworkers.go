package cleanupmanager

import (
	"context"
	"time"

	"github.com/riverqueue/river"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
)

const (
	QueueName = "engine-cleanup"
)

func AddWorkers(ctx context.Context, db gadb.DBTX, w *river.Workers) error {
	return river.AddWorkerSafely(w, WorkerOnCall(db))
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
				return OnCallArgs{MaxAgeDays: cfg.Maintenance.ScheduleCleanupDays}, &river.InsertOpts{
					Queue:       QueueName,
					MaxAttempts: 25,
				}
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		))
	}

	return nil
}
