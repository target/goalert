package cleanupmanager

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/alert/alertlog"
	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/util/jsonutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
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
		return fmt.Errorf("begin tx: %w", err)
	}
	defer sqlutil.Rollback(ctx, "cleanup manager", tx)

	_, err = tx.StmtContext(ctx, db.setTimeout).ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("set timeout: %w", err)
	}

	var now time.Time
	err = tx.StmtContext(ctx, db.now).QueryRowContext(ctx).Scan(&now)
	if err != nil {
		return err
	}

	_, err = tx.StmtContext(ctx, db.cleanupSessions).ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("cleanup sessions: %w", err)
	}

	cfg := config.FromContext(ctx)
	if cfg.Maintenance.AlertCleanupDays > 0 {
		var dur pgtype.Interval
		dur.Days = int32(cfg.Maintenance.AlertCleanupDays)
		dur.Status = pgtype.Present
		_, err = tx.StmtContext(ctx, db.cleanupAlerts).ExecContext(ctx, &dur)
		if err != nil {
			return fmt.Errorf("cleanup alerts: %w", err)
		}
	}

	if cfg.Maintenance.AlertAutoCloseDays > 0 {
		rows, err := tx.StmtContext(ctx, db.staleAlerts).QueryContext(ctx, cfg.Maintenance.AlertAutoCloseDays, cfg.Maintenance.AutoCloseAckedAlerts)
		if err != nil {
			return fmt.Errorf("query auto-close alerts: %w", err)
		}
		defer rows.Close()
		var ids []int
		for rows.Next() {
			var id int
			err = rows.Scan(&id)
			if err != nil {
				return fmt.Errorf("cleanup auto-close alerts: scan : %w", err)
			}
			ids = append(ids, id)
		}
		var autoCloseDays alertlog.AutoClose
		autoCloseDays.AlertAutoCloseDays = cfg.Maintenance.AlertAutoCloseDays
		_, err = db.alertStore.UpdateManyAlertStatus(ctx, alert.StatusClosed, ids, autoCloseDays)
		if err != nil {
			return fmt.Errorf("cleanup auto-close alerts: %w", err)
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
	if cfg.Maintenance.ScheduleCleanupDays > 0 {
		var dur pgtype.Interval
		dur.Days = int32(cfg.Maintenance.ScheduleCleanupDays)
		dur.Status = pgtype.Present
		_, err = tx.StmtContext(ctx, db.cleanupOverrides).ExecContext(ctx, &dur)
		if err != nil {
			return fmt.Errorf("cleanup overrides: %w", err)
		}

		_, err = tx.StmtContext(ctx, db.cleanupSchedOnCall).ExecContext(ctx, &dur)
		if err != nil {
			return fmt.Errorf("cleanup schedule on-call: %w", err)
		}

		_, err = tx.StmtContext(ctx, db.cleanupEPOnCall).ExecContext(ctx, &dur)
		if err != nil {
			return fmt.Errorf("cleanup escalation policy on-call: %w", err)
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
		Raw  json.RawMessage
	}
	var m []schedData
	for rows.Next() {
		var data schedData
		err = rows.Scan(&data.ID, &data.Raw)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data.Raw, &data.Data)
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
	schedCuttoff := now.AddDate(-1, 0, 0)
	for _, dat := range m {
		cleanupScheduleData(&dat.Data, lookup, schedCuttoff)
		rawData, err := jsonutil.Apply(dat.Raw, dat.Data)
		if err != nil {
			return err
		}
		_, err = tx.StmtContext(ctx, db.setSchedData).ExecContext(ctx, dat.ID, rawData)
		if err != nil {
			return fmt.Errorf("cleanup api keys: %w", err)
		}
	}

	err = tx.StmtContext(ctx, db.cleanupAlertLogs).QueryRowContext(ctx, db.logIndex).Scan(&db.logIndex)
	if errors.Is(err, sql.ErrNoRows) {
		// repeat
		db.logIndex = 0
		err = nil
	}
	if err != nil {
		return fmt.Errorf("cleanup alert_logs: %w", err)
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

func cleanupScheduleData(data *schedule.Data, userMap map[string]struct{}, cutoff time.Time) {
	filtered := data.V1.TemporarySchedules[:0]
	for _, temp := range data.V1.TemporarySchedules {
		if temp.End.Before(cutoff) {
			continue
		}
		filtered = append(filtered, temp)
	}
	data.V1.TemporarySchedules = filtered
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
