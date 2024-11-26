package engine

import (
	"context"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// scheduleAllPeriodicJobs schedules all periodic jobs immediately.
func (p *Engine) scheduleAllPeriodicJobs(ctx context.Context) error {
	var jobs []river.InsertManyParams
	for _, fn := range p.periodicJobs {
		args, opts := fn()
		jobs = append(jobs, river.InsertManyParams{
			Args:       args,
			InsertOpts: opts,
		})
	}

	_, err := p.cfg.River.InsertManyFast(ctx, jobs)
	return err
}

// waitForAllJobs waits for all existing jobs to complete.
func (p *Engine) waitForAllJobs(ctx context.Context) error {
	for {
		res, err := p.cfg.River.JobList(ctx, river.NewJobListParams().States(
			rivertype.JobStateAvailable,
			rivertype.JobStateRunning,
			rivertype.JobStateScheduled,
			rivertype.JobStatePending,
		))
		if err != nil {
			return fmt.Errorf("job list: %w", err)
		}
		if len(res.Jobs) == 0 {
			return nil
		}
	}
}
