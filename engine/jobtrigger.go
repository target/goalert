package engine

import (
	"context"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// runAllPeriodicJobs schedules all periodic jobs immediately, and waits for them to complete.
func (p *Engine) runAllPeriodicJobs(ctx context.Context) error {
	var jobs []river.InsertManyParams
	for _, fn := range p.periodicJobs {
		args, opts := fn()
		jobs = append(jobs, river.InsertManyParams{Args: args, InsertOpts: opts})
	}

	_, err := p.cfg.River.InsertManyFast(ctx, jobs)
	if err != nil {
		return err
	}

	for {
		res, err := p.cfg.River.JobList(ctx, river.NewJobListParams().States(rivertype.JobStateAvailable, rivertype.JobStateRunning, rivertype.JobStateScheduled, rivertype.JobStatePending))
		if err != nil {
			return fmt.Errorf("job list: %w", err)
		}
		if len(res.Jobs) == 0 {
			return nil
		}
	}
}
