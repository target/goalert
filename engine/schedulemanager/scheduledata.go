package schedulemanager

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/util/log"
)

func (db *DB) migrateScheduleDataNotifDedup(ctx context.Context, tx *sql.Tx) (bool, error) {
	gdb := gadb.New(tx)
	if db.migrateMap == nil {
		rows, err := gdb.SchedMgrNCDedupMapping(ctx)
		if err != nil {
			return false, fmt.Errorf("get schedule data notification dedup mapping: %w", err)
		}
		db.migrateMap = make(map[uuid.UUID]uuid.UUID, len(rows))
		for _, row := range rows {
			db.migrateMap[row.OldID] = row.NewID
		}
	}
	if db.migrateSchedIDs == nil {
		rows, err := gdb.SchedMgrDataIDs(ctx)
		if err != nil {
			return false, fmt.Errorf("get schedule data ids: %w", err)
		}
		if len(rows) == 0 {
			return true, nil
		}
		db.migrateSchedIDs = rows
	}

	t := time.Now()
	// process as many as we can in 5 seconds
	for len(db.migrateSchedIDs) > 0 {
		if time.Since(t) > 5*time.Second {
			return false, nil
		}

		sched := db.migrateSchedIDs[0]
		rawData, err := gdb.SchedMgrGetData(ctx, sched)
		if errors.Is(err, sql.ErrNoRows) {
			db.migrateSchedIDs = db.migrateSchedIDs[1:]
			continue
		}
		if err != nil {
			return false, fmt.Errorf("get schedule data: %w", err)
		}
		var data schedule.Data
		err = json.Unmarshal(rawData, &data)
		if err != nil {
			return false, fmt.Errorf("unmarshal schedule data: %w", err)
		}

		var hadUpdate bool
		for i, rule := range data.V1.OnCallNotificationRules {
			newID, ok := db.migrateMap[rule.ChannelID]
			if !ok {
				continue
			}
			data.V1.OnCallNotificationRules[i].ChannelID = newID
			hadUpdate = true
		}
		if !hadUpdate {
			db.migrateSchedIDs = db.migrateSchedIDs[1:]
			continue
		}

		// Note: We are passing the notification rules array directly instead of the entire data object because in the query, we're updating only that field.
		// This is done to ensure we don't overwrite/erase any other fields that may be present in the data object.
		newData, err := json.Marshal(data.V1.OnCallNotificationRules)
		if err != nil {
			return false, fmt.Errorf("marshal schedule data: %w", err)
		}

		log.Logf(ctx, "Migrating schedule data for schedule %s", sched)
		err = gdb.SchedMgrSetDataV1Rules(ctx, gadb.SchedMgrSetDataV1RulesParams{
			ScheduleID:  sched,
			Replacement: newData,
		})
		if err != nil {
			return false, fmt.Errorf("set schedule data: %w", err)
		}

		db.migrateSchedIDs = db.migrateSchedIDs[1:]
	}

	db.migrateMap = nil
	db.migrateSchedIDs = nil

	return true, nil
}
