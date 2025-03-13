package rotationmanager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/target/goalert/gadb"
)

type LookForWorkArgs struct{}

func (LookForWorkArgs) Kind() string { return "rotation-manager-lfw" }

// lookForWork will schedule jobs for rotations in the entity_updates table.
func (db *DB) lookForWork(ctx context.Context, j *river.Job[LookForWorkArgs]) error {
	var hadWork bool
	err := db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		g := gadb.New(tx)

		rotations, err := g.RotMgrFindWork(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			// done, no more work
			return nil
		}
		if err != nil {
			return fmt.Errorf("find work: %w", err)
		}
		if len(rotations) == 0 {
			return nil
		}

		var params []river.InsertManyParams
		for _, r := range rotations {
			params = append(params, river.InsertManyParams{
				Args: UpdateArgs{RotationID: r},
				InsertOpts: &river.InsertOpts{
					Queue:    QueueName,
					Priority: PriorityEvent,
				},
			})
		}

		_, err = db.riverDBSQL.InsertManyFastTx(ctx, tx, params)
		if err != nil {
			return fmt.Errorf("insert many: %w", err)
		}
		hadWork = true
		return nil
	})
	if err != nil {
		return fmt.Errorf("look for work: %w", err)
	}
	if !hadWork {
		return nil
	}

	// There was work to do, so wait a bit before looking again.
	return river.JobSnooze(time.Second)
}
