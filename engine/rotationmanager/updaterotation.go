package rotationmanager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/util"
)

type UpdateArgs struct {
	RotationID uuid.UUID
}

func (UpdateArgs) Kind() string { return "rotation-manager-update" }

// updateRotation updates the state of a single rotation, and schedules a job for the next rotation time.
func (db *DB) updateRotation(ctx context.Context, j *river.Job[UpdateArgs]) error {
	return db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		g := gadb.New(tx)

		row, err := g.RotMgrRotationData(ctx, j.Args.RotationID)
		if errors.Is(err, sql.ErrNoRows) {
			// no longer exists, so nothing to do
			return nil
		}
		if err != nil {
			return fmt.Errorf("load rotation data: %w", err)
		}

		if len(row.Participants) == 0 {
			if row.StateVersion != 0 {
				// no participants, but we have a position, so clear it
				err = g.RotMgrEnd(ctx, j.Args.RotationID)
				if err != nil {
					return fmt.Errorf("end rotation: %w", err)
				}
				return nil
			}

			return nil
		}

		loc, err := util.LoadLocation(row.Rotation.TimeZone)
		if err != nil {
			return fmt.Errorf("load location: %w", err)
		}

		r := rotation.Rotation{
			Type:        rotation.Type(row.Rotation.Type),
			Start:       row.Rotation.StartTime.In(loc),
			ShiftLength: int(row.Rotation.ShiftLength),
		}

		// schedule next run
		_, err = db.riverDBSQL.InsertTx(ctx, tx, UpdateArgs{RotationID: j.Args.RotationID}, &river.InsertOpts{
			UniqueOpts: river.UniqueOpts{
				ByArgs:   true,
				ByPeriod: time.Minute,
			},
			Priority:    PriorityScheduled,
			ScheduledAt: r.EndTime(row.Now),
			Queue:       QueueName,
		})
		if err != nil {
			return fmt.Errorf("schedule next run: %w", err)
		}

		if row.StateVersion == 0 {
			// no state, but we have participants, so start at the beginning
			err = g.RotMgrStart(ctx, j.Args.RotationID)
			if err != nil {
				return fmt.Errorf("start rotation: %w", err)
			}
			return nil
		}

		s := rotState{
			ShiftStart: row.StateShiftStart.Time.In(loc),
			Position:   int(row.StatePosition),
			Version:    int(row.StateVersion),
		}
		adv, err := calcAdvance(ctx, row.Now, &r, s, len(row.Participants))
		if err != nil {
			return fmt.Errorf("calc advance: %w", err)
		}
		if adv == nil {
			// no advancement needed
			return nil
		}

		err = g.RotMgrUpdate(ctx, gadb.RotMgrUpdateParams{
			RotationID:            j.Args.RotationID,
			Position:              int32(adv.newPosition),
			RotationParticipantID: row.Participants[adv.newPosition],
		})
		if err != nil {
			return fmt.Errorf("update rotation state (advance): %w", err)
		}

		return nil
	})
}
