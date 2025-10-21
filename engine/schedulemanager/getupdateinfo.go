package schedulemanager

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/util"
)

// getUpdateInfo retrieves all necessary information about a schedule to calculate updates.
func getUpdateInfo(ctx context.Context, tx *sql.Tx, scheduleID uuid.UUID) (*updateInfo, error) {
	info := updateInfo{
		ScheduleID:    scheduleID,
		CurrentOnCall: mapset.NewThreadUnsafeSet[uuid.UUID](),
	}
	info.ScheduleID = scheduleID
	db := gadb.New(tx)
	tz, err := db.SchedMgrTimezone(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("lookup timezone: %w", err)
	}
	info.TimeZone, err = util.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("load location: %w", err)
	}
	info.RawScheduleData, err = db.SchedFindDataForUpdate(ctx, scheduleID)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		return nil, fmt.Errorf("lookup schedule data: %w", err)
	}
	if info.RawScheduleData != nil {
		err = json.Unmarshal(info.RawScheduleData, &info.ScheduleData)
	}
	if err != nil {
		return nil, fmt.Errorf("unmarshal schedule data: %w", err)
	}
	onCallIDs, err := db.SchedMgrOnCall(ctx, scheduleID)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		return nil, fmt.Errorf("lookup on-call users: %w", err)
	}
	for _, id := range onCallIDs {
		info.CurrentOnCall.Add(id)
	}
	info.Rules, err = db.SchedMgrRules(ctx, scheduleID)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		return nil, fmt.Errorf("lookup schedule rules: %w", err)
	}
	info.Overrides, err = db.SchedMgrOverrides(ctx, scheduleID)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		return nil, fmt.Errorf("lookup active overrides: %w", err)
	}

	return &info, nil
}
