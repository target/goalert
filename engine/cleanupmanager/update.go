package cleanupmanager

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jackc/pgtype"
	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/util/log"
)

// UpdateAll will update the state of all active escalation policies.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := db.update(ctx)
	return err
}

func (db *DB) update(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	log.Debugf(ctx, "Running cleanup operations.")

	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.StmtContext(ctx, db.setTimeout).ExecContext(ctx)
	if err != nil {
		return err
	}

	var now time.Time
	err = tx.StmtContext(ctx, db.now).QueryRowContext(ctx).Scan(&now)
	if err != nil {
		return err
	}

	_, err = tx.StmtContext(ctx, db.cleanupSessions).ExecContext(ctx)
	if err != nil {
		return err
	}

	cfg := config.FromContext(ctx)
	if cfg.Maintenance.AlertCleanupDays > 0 {
		var dur pgtype.Interval
		dur.Days = int32(cfg.Maintenance.AlertCleanupDays)
		dur.Status = pgtype.Present
		_, err = tx.StmtContext(ctx, db.cleanupAlerts).ExecContext(ctx, &dur)
		if err != nil {
			return err
		}
	}
	if cfg.Maintenance.APIKeyExpireDays > 0 {
		var dur pgtype.Interval
		dur.Days = int32(cfg.Maintenance.APIKeyExpireDays)
		dur.Status = pgtype.Present
		_, err = tx.StmtContext(ctx, db.cleanupAPIKeys).ExecContext(ctx, &dur)
		if err != nil {
			return err
		}
	}

	rows, err := tx.StmtContext(ctx, db.schedData).QueryContext(ctx)
	if err != nil {
		return err
	}
	defer rows.Close()

	type schedData struct {
		ID   string
		Data schedule.Data
	}
	var m []schedData
	for rows.Next() {
		var data schedData
		var rawData json.RawMessage
		err = rows.Scan(&data.ID, &rawData)
		if err != nil {
			return err
		}
		err = json.Unmarshal(rawData, &data.Data)
		if err != nil {
			return err
		}
		m = append(m, data)
	}
	var currentUsers []string
	if len(m) > 0 {
		currentUsers, err = db.getUsers(ctx, tx)
		if err != nil {
			return err
		}
	}

	lookup := lookupMap(currentUsers)
	for _, dat := range m {
		cleanupScheduleData(&dat.Data, lookup, now)
		rawData, err := json.Marshal(dat.Data)
		if err != nil {
			return err
		}
		_, err = tx.StmtContext(ctx, db.setSchedData).ExecContext(ctx, dat.ID, rawData)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func lookupMap(users []string) map[string]struct{} {
	userLookup := make(map[string]struct{}, len(users))
	for _, id := range users {
		userLookup[id] = struct{}{}
	}
	return userLookup
}

func cleanupScheduleData(data *schedule.Data, userMap map[string]struct{}, now time.Time) {
	for idx, grp := range data.V1.TemporarySchedules {
		shifts := grp.Shifts
		grp.Shifts = grp.Shifts[:0]
		for _, shift := range shifts {
			if _, ok := userMap[shift.UserID]; !ok {
				continue
			}
			grp.Shifts = append(grp.Shifts, shift)
		}

		data.V1.TemporarySchedules[idx] = schedule.TrimGroupStart(grp, now)
	}

	data.V1.TemporarySchedules = schedule.MergeGroups(data.V1.TemporarySchedules)
}

// getUsers retrieves the current set of user IDs
func (db *DB) getUsers(ctx context.Context, tx *sql.Tx) ([]string, error) {
	rows, err := tx.StmtContext(ctx, db.userIDs).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if err == sql.ErrNoRows {
		return nil, nil
	}

	var users []string
	var id string
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		users = append(users, id)
	}

	return users, nil
}
