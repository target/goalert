package graphql

import (
	"context"
	"database/sql"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"
)

type legacyDB struct {
	db *sql.DB

	schedFromRot    *sql.Stmt
	rotFromSched    *sql.Stmt
	allRotFromSched *sql.Stmt
}

func newLegacyDB(ctx context.Context, db *sql.DB) (*legacyDB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &legacyDB{
		db: db,

		schedFromRot: p.P(`
			SELECT schedule_id
			FROM schedule_rules
			WHERE tgt_rotation_id = $1 and tgt_rotation_id notnull
			LIMIT 1
		`),
		rotFromSched: p.P(`
			SELECT tgt_rotation_id
			FROM schedule_rules
			WHERE schedule_id = $1 and tgt_rotation_id notnull
			LIMIT 1
		`),
		allRotFromSched: p.P(`
			SELECT DISTINCT tgt_rotation_id
			FROM schedule_rules
			WHERE schedule_id = $1 and tgt_rotation_id notnull
		`),
	}, p.Err
}
func (l *legacyDB) ScheduleIDFromRotation(ctx context.Context, rotID string) (string, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return "", err
	}
	err = validate.UUID("RotationID", rotID)
	if err != nil {
		return "", err
	}

	var schedID string
	err = l.schedFromRot.QueryRowContext(ctx, rotID).Scan(&schedID)
	if err != nil {
		return "", err
	}
	return schedID, nil
}

func (l *legacyDB) RotationIDFromScheduleID(ctx context.Context, schedID string) (string, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return "", err
	}
	err = validate.UUID("ScheduleID", schedID)
	if err != nil {
		return "", err
	}

	var rotID string
	err = l.rotFromSched.QueryRowContext(ctx, schedID).Scan(&rotID)
	if err != nil {
		return "", err
	}
	return rotID, nil
}
func (l *legacyDB) FindAllRotationIDsFromScheduleID(ctx context.Context, schedID string) ([]string, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ScheduleID", schedID)
	if err != nil {
		return nil, err
	}

	rows, err := l.allRotFromSched.QueryContext(ctx, schedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rotationIDs []string
	for rows.Next() {
		var rotID string
		err = rows.Scan(&rotID)
		if err != nil {
			return nil, err
		}
		rotationIDs = append(rotationIDs, rotID)
	}
	return rotationIDs, nil
}
