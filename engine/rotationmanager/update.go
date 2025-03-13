package rotationmanager

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

// UpdateAll will update and cleanup the rotation state for all rotations.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	err = db.update(ctx, true, nil)
	return err
}

// UpdateOneRotation will update and cleanup the rotation state for the given rotation.
func (db *DB) UpdateOneRotation(ctx context.Context, rotID string) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	err = validate.UUID("Rotation", rotID)
	if err != nil {
		return err
	}
	ctx = log.WithField(ctx, "RotationID", rotID)
	return db.update(ctx, false, &rotID)
}

func (db *DB) update(ctx context.Context, all bool, rotID *string) error {
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

	_, err = tx.StmtContext(ctx, db.lockPart).ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "lock rotation participants")
	}

	needsAdvance, err := db.calcAdvances(ctx, tx, all, rotID)
	if err != nil {
		return errors.Wrap(err, "calc stale rotations")
	}

	updateStmt := tx.Stmt(db.rotate)
	for _, adv := range needsAdvance {
		fctx := log.WithFields(ctx, log.Fields{
			"RotationID": adv.id,
			"Position":   adv.newPosition,
		})

		if !adv.silent {
			log.Debugf(fctx, "Advancing rotation.")
		}
		_, err = updateStmt.ExecContext(fctx, adv.id, adv.newPosition)
		if err != nil {
			return errors.Wrap(err, "advance rotation")
		}
	}

	return errors.Wrap(tx.Commit(), "commit transaction")
}

func (db *DB) calcAdvances(ctx context.Context, tx *sql.Tx, all bool, rotID *string) ([]advance, error) {
	var t time.Time
	err := tx.Stmt(db.currentTime).QueryRowContext(ctx).Scan(&t)
	if err != nil {
		return nil, errors.Wrap(err, "fetch current timestamp")
	}

	rows, err := tx.Stmt(db.rotateData).QueryContext(ctx, all, rotID)
	if err != nil {
		return nil, errors.Wrap(err, "fetch current rotation state")
	}
	defer rows.Close()

	var rot rotation.Rotation
	var state rotState
	var partCount int
	var tzName string
	var adv *advance
	var loc *time.Location
	var needsAdvance []advance

	for rows.Next() {
		err = rows.Scan(
			&rot.ID,
			&rot.Type,
			&rot.Start,
			&rot.ShiftLength,
			&tzName,
			&state.ShiftStart,
			&state.Position,
			&partCount,
			&state.Version,
		)
		if err != nil {
			return nil, errors.Wrap(err, "scan rotation data")
		}
		loc, err = util.LoadLocation(tzName)
		if err != nil {
			return nil, errors.Wrap(err, "load timezone")
		}
		rot.Start = rot.Start.In(loc)
		adv, err = calcAdvance(ctx, t, &rot, state, partCount)
		if err != nil {
			return nil, errors.Wrap(err, "calculate rotation advance")
		}
		if adv != nil {
			needsAdvance = append(needsAdvance, *adv)
			if len(needsAdvance) == 150 {
				// only process up to 150 at a time (of those that need updates)
				break
			}
		}
	}
	return needsAdvance, nil
}
