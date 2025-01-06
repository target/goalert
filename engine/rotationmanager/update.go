package rotationmanager

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

// UpdateAll will update and cleanup the rotation state for all rotations.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	err = db.update(ctx)
	return err
}

func (db *DB) update(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	log.Debugf(ctx, "Updating rotations.")

	// process rotation advancement
	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "start advancement transaction")
	}
	defer sqlutil.Rollback(ctx, "rotation manager", tx)

	gdb := gadb.New(tx)

	err = gdb.RotationMgrLock(ctx)
	if err != nil {
		return errors.Wrap(err, "lock rotation participants")
	}

	now, err := gdb.Now(ctx)
	if err != nil {
		return errors.Wrap(err, "fetch current timestamp")
	}
	rows, err := gdb.RotationMgrGetConfig(ctx)
	if err != nil {
		return errors.Wrap(err, "fetch rotation configs")
	}
	needsAdvance, err := db.calcAdvances(ctx, rows, now)
	if err != nil {
		return errors.Wrap(err, "calc stale rotations")
	}

	for _, adv := range needsAdvance {
		fctx := log.WithFields(ctx, log.Fields{
			"RotationID": adv.id,
			"Position":   adv.newPosition,
		})

		if !adv.silent {
			log.Debugf(fctx, "Advancing rotation.")
		}
		err = gdb.RotationMgrUpdateState(ctx, gadb.RotationMgrUpdateStateParams{
			RotationID: adv.id,
			Position:   int32(adv.newPosition),
		})
		if err != nil {
			return errors.Wrap(err, "advance rotation")
		}
	}

	return errors.Wrap(tx.Commit(), "commit transaction")
}

func (db *DB) calcAdvances(ctx context.Context, rows []gadb.RotationMgrGetConfigRow, t time.Time) ([]advance, error) {
	var needsAdvance []advance
	for _, row := range rows {
		rot := rotation.Rotation{
			ID:          row.ID.String(),
			Type:        rotation.Type(row.Type),
			ShiftLength: int(row.ShiftLength),
		}
		loc, err := util.LoadLocation(row.TimeZone)
		if err != nil {
			return nil, errors.Wrap(err, "load timezone")
		}
		rot.Start = row.StartTime.In(loc)
		state := rotState{
			State: rotation.State{
				ShiftStart: row.ShiftStart,
				Position:   int(row.Position),
			},
			Version: int(row.Version),
		}
		adv := calcAdvance(ctx, t, row.ID, &rot, state, int(row.ParticipantCount))
		if adv == nil {
			continue
		}

		needsAdvance = append(needsAdvance, *adv)
		if len(needsAdvance) == 150 {
			// only process up to 150 at a time (of those that need updates)
			break
		}
	}

	return needsAdvance, nil
}
