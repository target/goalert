package cleanupmanager

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/alert/alertlog"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
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

	q := gadb.New(tx)

	err = q.CleanMgrSetTimeout(ctx)
	if err != nil {
		return fmt.Errorf("set timeout: %w", err)
	}

	now, err := q.Now(ctx)
	if err != nil {
		return err
	}

	_, err = q.CleanMgrUserSessions(ctx, now.AddDate(0, 0, -30))
	if err != nil {
		return fmt.Errorf("cleanup sessions: %w", err)
	}

	cfg := config.FromContext(ctx)
	if cfg.Maintenance.AlertCleanupDays > 0 {
		_, err = q.CleanMgrClosedAlerts(ctx, now.AddDate(0, 0, -cfg.Maintenance.AlertCleanupDays))
		if err != nil {
			return fmt.Errorf("cleanup alerts: %w", err)
		}
	}

	if cfg.Maintenance.AlertAutoCloseDays > 0 {
		alertIDs, err := q.CleanMgrStaleAlertIDs(ctx, gadb.CleanMgrStaleAlertIDsParams{
			Cutoff:        now.AddDate(0, 0, -cfg.Maintenance.AlertAutoCloseDays),
			IncludeActive: cfg.Maintenance.AutoCloseAckedAlerts,
		})
		if err != nil {
			return fmt.Errorf("query auto-close alerts: %w", err)
		}
		ids := make([]int, len(alertIDs))
		for i, id := range alertIDs {
			ids[i] = int(id)
		}
		var autoCloseDays alertlog.AutoClose
		autoCloseDays.AlertAutoCloseDays = cfg.Maintenance.AlertAutoCloseDays
		_, err = db.alertStore.UpdateManyAlertStatus(ctx, alert.StatusClosed, ids, autoCloseDays)
		if err != nil {
			return fmt.Errorf("cleanup auto-close alerts: %w", err)
		}
	}

	if cfg.Maintenance.APIKeyExpireDays > 0 {
		_, err = q.CleanMgrCalSubs(ctx, now.AddDate(0, 0, -cfg.Maintenance.APIKeyExpireDays))
		if err != nil {
			return err
		}
	}
	if cfg.Maintenance.ScheduleCleanupDays > 0 {
		cutoff := now.AddDate(0, 0, -cfg.Maintenance.ScheduleCleanupDays)
		_, err = q.CleanMgrUserOverrides(ctx, cutoff)
		if err != nil {
			return fmt.Errorf("cleanup overrides: %w", err)
		}

		_, err = q.CleanMgrSchedHistory(ctx, cutoff)
		if err != nil {
			return fmt.Errorf("cleanup schedule on-call: %w", err)
		}

		_, err := q.CleanMgrEPHistory(ctx, cutoff)
		if err != nil {
			return fmt.Errorf("cleanup escalation policy on-call: %w", err)
		}
	}

	scheduleDataRows, err := q.CleanMgrScheduleData(ctx, now.AddDate(0, -1, 0))
	if err != nil {
		return err
	}

	var currentUsers []uuid.UUID
	if len(scheduleDataRows) > 0 {
		currentUsers, err = q.CleanMgrUserIDs(ctx)
		if err != nil {
			return err
		}
	}
	lookup := lookupMap(currentUsers)
	schedCuttoff := now.AddDate(-1, 0, 0)
	for _, row := range scheduleDataRows {
		var data schedule.Data
		err := json.Unmarshal(row.Data, &data)
		if err != nil {
			return fmt.Errorf("unmarshal schedule data %s: %w", row.ScheduleID.String(), err)
		}

		cleanupScheduleData(&data, lookup, schedCuttoff)
		rawData, err := jsonutil.Apply(row.Data, data)
		if err != nil {
			return err
		}
		err = q.CleanMgrUpdateScheduleData(ctx, gadb.CleanMgrUpdateScheduleDataParams{
			ScheduleID: row.ScheduleID,
			Data:       rawData,
		})
		if err != nil {
			return fmt.Errorf("cleanup api keys: %w", err)
		}
	}

	newIndex, err := q.CleanMgrAlertLogs(ctx, db.logIndex)
	if errors.Is(err, sql.ErrNoRows) {
		// repeat
		db.logIndex = 0
		err = nil
	}
	if err != nil {
		return fmt.Errorf("cleanup alert_logs: %w", err)
	}
	db.logIndex = newIndex

	return tx.Commit()
}

func lookupMap(users []uuid.UUID) map[uuid.UUID]struct{} {
	userLookup := make(map[uuid.UUID]struct{}, len(users))
	for _, id := range users {
		userLookup[id] = struct{}{}
	}
	return userLookup
}

func cleanupScheduleData(data *schedule.Data, userMap map[uuid.UUID]struct{}, cutoff time.Time) {
	filtered := data.V1.TemporarySchedules[:0]
	for _, temp := range data.V1.TemporarySchedules {
		if temp.End.Before(cutoff) {
			continue
		}
		filtered = append(filtered, temp)
	}
	data.V1.TemporarySchedules = filtered
}
