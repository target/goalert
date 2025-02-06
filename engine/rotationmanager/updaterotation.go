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

// cleanup is a worker function that will remove any stale subscriptions.
func (db *DB) updateRotation(ctx context.Context, j *river.Job[UpdateArgs]) error {
	return db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		g := gadb.New(tx)

		row, err := g.RotMgrRotationData(ctx, j.Args.RotationID)
		if errors.Is(err, sql.ErrNoRows) {
			// no longer exists, so nothing to do
			return nil
		}
		if err != nil {
			return err
		}

		if len(row.Participants) == 0 {
			if row.Position.Valid {
				// no participants, but we have a position, so clear it
				return g.RotMgrEnd(ctx, j.Args.RotationID)
			}

			return nil
		}

		loc, err := util.LoadLocation(row.TimeZone)
		if err != nil {
			return fmt.Errorf("load location: %w", err)
		}

		r := rotation.Rotation{
			Type:        rotation.Type(row.Type),
			Start:       row.StartTime.In(loc),
			ShiftLength: int(row.ShiftLength),
		}

		// schedule next run
		_, err = db.riverDBSQL.InsertTx(ctx, tx, UpdateArgs{RotationID: j.Args.RotationID}, &river.InsertOpts{
			UniqueOpts: river.UniqueOpts{
				ByArgs:   true,
				ByPeriod: time.Minute,
			},
			Priority:    PriorityScheduled,
			ScheduledAt: r.EndTime(row.Now),
		})
		if err != nil {
			return fmt.Errorf("schedule next run: %w", err)
		}

		if !row.Position.Valid {
			// no state, but we have participants, so start at the beginning
			return g.RotMgrStart(ctx, j.Args.RotationID)
		}

		s := rotState{
			ShiftStart: row.ShiftStart.Time.In(loc),
			Position:   int(row.Position.Int32),
		}
		adv := calcAdvance(ctx, row.Now, &r, s, len(row.Participants))
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
