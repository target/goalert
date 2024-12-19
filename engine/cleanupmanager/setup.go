package cleanupmanager

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

var _ processinglock.Setupable = &DB{}

const QueueName = "cleanup-manager"

const (
	PriorityAlertCleanup = 1
	PrioritySchedHistory = 1
	PriorityTempSchedLFW = 2
	PriorityAlertLogsLFW = 2
	PriorityTempSched    = 3
	PriorityAlertLogs    = 4
)

// whileWork will run the provided function in a loop until it returns done=true.
func (db *DB) whileWork(ctx context.Context, run func(ctx context.Context, tx *sql.Tx) (done bool, err error)) error {
	var done bool
	for {
		err := db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
			var err error
			done, err = run(ctx, tx)
			return err
		})
		if err != nil {
			return fmt.Errorf("do work: %w", err)
		}
		if done {
			break
		}

		err = util.ContextSleep(ctx, 100*time.Millisecond)
		if err != nil {
			return fmt.Errorf("sleep: %w", err)
		}
	}

	return nil
}

// Setup implements processinglock.Setupable.
func (db *DB) Setup(ctx context.Context, args processinglock.SetupArgs) error {
	river.AddWorker(args.Workers, river.WorkFunc(db.CleanupAlerts))
	river.AddWorker(args.Workers, river.WorkFunc(db.CleanupShifts))
	river.AddWorker(args.Workers, river.WorkFunc(db.CleanupScheduleData))
	river.AddWorker(args.Workers, river.WorkFunc(db.LookForWorkScheduleData))
	river.AddWorker(args.Workers, river.WorkFunc(db.CleanupAlertLogs))
	river.AddWorker(args.Workers, river.WorkFunc(db.LookForWorkAlertLogs))

	err := args.River.Queues().Add(QueueName, river.QueueConfig{MaxWorkers: 5})
	if err != nil {
		return fmt.Errorf("add queue: %w", err)
	}

	args.River.PeriodicJobs().AddMany([]*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return AlertArgs{}, &river.InsertOpts{
					Queue:    QueueName,
					Priority: PriorityAlertCleanup,
				}
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	})

	args.River.PeriodicJobs().AddMany([]*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(24*time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return ShiftArgs{}, &river.InsertOpts{
					Queue:    QueueName,
					Priority: PrioritySchedHistory,
				}
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	})

	args.River.PeriodicJobs().AddMany([]*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(24*time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return SchedDataLFW{}, &river.InsertOpts{
					Queue:    QueueName,
					Priority: PriorityTempSchedLFW,
				}
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	})

	args.River.PeriodicJobs().AddMany([]*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(7*24*time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return AlertLogLFWArgs{}, &river.InsertOpts{
					Queue: QueueName,
					UniqueOpts: river.UniqueOpts{
						ByArgs: true,
					},
				}
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	})

	return nil
}
