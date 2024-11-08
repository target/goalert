package cleanupmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
)

type OnCallArgs struct {
	MaxAgeDays int
}

func (OnCallArgs) Kind() string { return "engine-cleanup-on-call" }

type trimWork struct {
	QueryFn func(context.Context, time.Time) (int64, error)
	MaxDays int
}

func GetWork(ctx context.Context, db *gadb.Queries) (work []trimWork) {
	cfg := config.FromContext(ctx)

	return []trimWork{
		// schedule stuff
		{db.CleanMgrTrimOverrides, cfg.Maintenance.ScheduleCleanupDays},
		{db.CleanMgrTrimSchedOnCall, cfg.Maintenance.ScheduleCleanupDays},
		{db.CleanMgrTrimEPOnCall, cfg.Maintenance.ScheduleCleanupDays},

		// alert stuff
		{db.CleanMgrTrimClosedAlerts, cfg.Maintenance.AlertCleanupDays},
	}
}

func TrimOldRows(queryFn func(context.Context, time.Time) (int64, error), maxDays int) river.Worker {
	return river.WorkFunc(func(ctx context.Context, job *river.Job[OnCallArgs]) error {
		now := time.Now()
		cutoff := now.AddDate(0, 0, -job.Args.MaxAgeDays)

		err := runWhileWork(ctx, func() (int64, error) { return queryFn(ctx, cutoff) })
		if err != nil {
			return fmt.Errorf("delete expired rows: %w", err)
		}

		return nil
	})
}

func WorkerOnCall(dbtx gadb.DBTX) river.Worker[OnCallArgs] {
	db := gadb.New(dbtx)
	return river.WorkFunc(func(ctx context.Context, job *river.Job[OnCallArgs]) error {
		now, err := db.Now(ctx)
		if err != nil {
			return fmt.Errorf("get current DB time: %w", err)
		}

		cutoff := now.AddDate(0, 0, -job.Args.MaxAgeDays)

		err = runWhileWork(ctx, func() (int64, error) { return db.CleanMgrTrimOverrides(ctx, cutoff) })
		if err != nil {
			return fmt.Errorf("trim overrides: %w", err)
		}

		err = runWhileWork(ctx, func() (int64, error) { return db.CleanMgrTrimSchedOnCall(ctx, cutoff) })
		if err != nil {
			return fmt.Errorf("trim schedule on-call: %w", err)
		}

		err = runWhileWork(ctx, func() (int64, error) { return db.CleanMgrTrimEPOnCall(ctx, cutoff) })
		if err != nil {
			return fmt.Errorf("trim escalation policy on-call: %w", err)
		}

		return nil
	})
}
