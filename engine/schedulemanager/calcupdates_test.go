package schedulemanager

import (
	"encoding/json"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/util/timeutil"
)

func TestUpdateInfo_calcUpdates_NotifyOnChange(t *testing.T) {
	channelID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	var sData schedule.Data
	sData.V1.OnCallNotificationRules = []schedule.OnCallNotificationRule{
		{
			ChannelID: channelID,
		},
	}
	data, err := json.Marshal(sData)
	require.NoError(t, err)
	info := updateInfo{
		ScheduleID:      uuid.New(),
		TimeZone:        time.UTC,
		RawScheduleData: data,
		ScheduleData:    sData,
		CurrentOnCall:   mapset.NewThreadUnsafeSet(uuid.New()), // no rules, so no longer on-call
		Rules:           []gadb.SchedMgrRulesRow{},
		ActiveOverrides: []gadb.SchedMgrOverridesRow{},
	}

	result, err := info.calcUpdates(time.Now())
	require.NoError(t, err)
	require.EqualValues(t, []uuid.UUID{channelID}, result.NotificationChannels.ToSlice())
}

func TestUpdateInfo_calcUpdates_NotifyAtTime(t *testing.T) {
	channelID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	var sData schedule.Data
	everyDay := timeutil.EveryDay()
	at9am := timeutil.NewClock(9, 0)
	nextRun := time.Date(2023, 10, 1, 9, 0, 0, 0, time.UTC)
	sData.V1.OnCallNotificationRules = []schedule.OnCallNotificationRule{
		{
			ChannelID:        channelID,
			NextNotification: &nextRun,
			WeekdayFilter:    &everyDay,
			Time:             &at9am,
		},
	}
	data, err := json.Marshal(sData)
	require.NoError(t, err)
	info := updateInfo{
		ScheduleID:      uuid.New(),
		TimeZone:        time.UTC,
		RawScheduleData: data,
		ScheduleData:    sData,
		CurrentOnCall:   mapset.NewThreadUnsafeSet[uuid.UUID](),
		Rules:           []gadb.SchedMgrRulesRow{},
		ActiveOverrides: []gadb.SchedMgrOverridesRow{},
	}

	result, err := info.calcUpdates(nextRun.Add(-1 * time.Hour))
	require.NoError(t, err)
	require.Empty(t, result.NotificationChannels.ToSlice()) // 8 am, no notification yet

	require.Empty(t, result.NewRawScheduleData) // no update necessary

	result, err = info.calcUpdates(nextRun)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{channelID}, result.NotificationChannels.ToSlice())

	var expected schedule.Data
	expectedNext := nextRun.AddDate(0, 0, 1)
	expected.V1.OnCallNotificationRules = []schedule.OnCallNotificationRule{
		{
			ChannelID:        channelID,
			NextNotification: &expectedNext,
			WeekdayFilter:    &everyDay,
			Time:             &at9am,
		},
	}
	expectedData, err := json.Marshal(expected)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedData), string(result.NewRawScheduleData))
}
