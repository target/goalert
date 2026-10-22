package schedulemanager

import (
	"encoding/json"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/util/timeutil"
)

func TestUpdateInfo_calcUpdates_NotifyOnChange(t *testing.T) {
	channelID1 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	channelID2 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")
	var sData schedule.Data
	sData.V1.OnCallNotificationRules = []schedule.OnCallNotificationRule{
		{ChannelID: channelID1}, {ChannelID: channelID2},
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
		Overrides:       []overrideSpan{},
	}

	result, err := info.calcUpdates(time.Now())
	require.NoError(t, err)
	require.ElementsMatch(t, []uuid.UUID{channelID1, channelID2}, result.NotificationChannels.ToSlice())
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
		Overrides:       []overrideSpan{},
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

func TestUpdateInfo_calcUpdates_TempSchedMissingUser(t *testing.T) {
	channelID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	validUser := uuid.MustParse("223e4567-e89b-12d3-a456-426614174000")
	deletedUser := uuid.MustParse("323e4567-e89b-12d3-a456-426614174000")
	now := time.Date(2023, 10, 1, 9, 0, 0, 0, time.UTC)

	var sData schedule.Data
	sData.V1.OnCallNotificationRules = []schedule.OnCallNotificationRule{{ChannelID: channelID}}
	sData.V1.TemporarySchedules = []schedule.TemporarySchedule{{
		Start: now.Add(-time.Hour),
		End:   now.Add(time.Hour),
		Shifts: []schedule.FixedShift{
			{Start: now.Add(-time.Hour), End: now.Add(time.Hour), UserID: validUser.String()},
			{Start: now.Add(-time.Hour), End: now.Add(time.Hour), UserID: deletedUser.String()},
		},
	}}
	data, err := json.Marshal(sData)
	require.NoError(t, err)

	info := updateInfo{
		ScheduleID:      uuid.New(),
		TimeZone:        time.UTC,
		RawScheduleData: data,
		ScheduleData:    sData,
		// the deleted user can never be recorded as on-call, so only the valid user is
		CurrentOnCall: mapset.NewThreadUnsafeSet(validUser),
		MissingUsers:  mapset.NewThreadUnsafeSet(deletedUser),
	}

	result, err := info.calcUpdates(now)
	require.NoError(t, err)
	require.Empty(t, result.UsersToStart.ToSlice(), "deleted user should not be started")
	require.Empty(t, result.UsersToStop.ToSlice())
	require.Empty(t, result.NotificationChannels.ToSlice(), "should not notify when on-call is unchanged")

	// without MissingUsers, the deleted user is always "new", which triggers a notification every cycle
	info.MissingUsers = nil
	result, err = info.calcUpdates(now)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{deletedUser}, result.UsersToStart.ToSlice())
	require.Equal(t, []uuid.UUID{channelID}, result.NotificationChannels.ToSlice())
}

func TestOverrideActive(t *testing.T) {
	start := time.Date(2023, 10, 1, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	o := overrideSpan{Start: start, End: end}

	check := func(desc string, now time.Time, exp bool) {
		t.Helper()
		assert.Equal(t, exp, overrideActive(now, o), desc)
	}

	// inclusive on both ends, matching `now() BETWEEN start_time AND end_time`
	check("before start", start.Add(-time.Nanosecond), false)
	check("at start", start, true)
	check("between", start.Add(30*time.Minute), true)
	check("at end", end, true)
	check("after end", end.Add(time.Nanosecond), false)
	check("other zone, same instant", start.In(time.FixedZone("x", -3600)), true)
}

func TestOverrideSpanFromRow(t *testing.T) {
	add := uuid.NullUUID{UUID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), Valid: true}
	rm := uuid.NullUUID{UUID: uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"), Valid: true}
	start := time.Date(2023, 10, 1, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	exp := overrideSpan{AddUserID: add, RemoveUserID: rm, Start: start, End: end}

	require.Equal(t, exp, overrideSpanFromRow(gadb.SchedMgrOverridesRow{
		AddUserID:     add,
		RemoveUserID:  rm,
		TgtScheduleID: uuid.New(),
		StartTime:     start,
		EndTime:       end,
	}))
	require.Equal(t, exp, overrideSpanFromScheduleRow(gadb.SchedMgrOverridesForScheduleRow{
		AddUserID:     add,
		RemoveUserID:  rm,
		TgtScheduleID: uuid.New(),
		StartTime:     start,
		EndTime:       end,
	}))
}

func TestUpdateInfo_calcLatestOnCall_OverrideWindow(t *testing.T) {
	now := time.Date(2023, 10, 1, 9, 0, 0, 0, time.UTC)
	ruleUser := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	futureUser := uuid.MustParse("223e4567-e89b-12d3-a456-426614174000")
	pastUser := uuid.MustParse("323e4567-e89b-12d3-a456-426614174000")
	activeUser := uuid.MustParse("423e4567-e89b-12d3-a456-426614174000")
	valid := func(id uuid.UUID) uuid.NullUUID { return uuid.NullUUID{UUID: id, Valid: true} }

	info := updateInfo{
		ScheduleID: uuid.New(),
		TimeZone:   time.UTC,
		// every day with Start == End is always active
		Rules: []gadb.SchedMgrRulesRow{{
			Sunday: true, Monday: true, Tuesday: true, Wednesday: true, Thursday: true, Friday: true, Saturday: true,
			ResolvedUserID: ruleUser,
		}},
		Overrides: []overrideSpan{
			// not yet started: must have no effect, even though it would replace ruleUser
			{AddUserID: valid(futureUser), RemoveUserID: valid(ruleUser), Start: now.Add(time.Minute), End: now.Add(time.Hour)},
			// already ended: must have no effect
			{AddUserID: valid(pastUser), Start: now.Add(-time.Hour), End: now.Add(-time.Minute)},
			// active: applies
			{AddUserID: valid(activeUser), Start: now.Add(-time.Minute), End: now.Add(time.Minute)},
		},
	}

	onCall := info.calcLatestOnCall(now)
	require.ElementsMatch(t, []uuid.UUID{ruleUser, activeUser}, onCall.ToSlice())

	// one minute later: the replace override starts (inclusive) and the
	// active one reaches its End (still inclusive)
	onCall = info.calcLatestOnCall(now.Add(time.Minute))
	require.ElementsMatch(t, []uuid.UUID{futureUser, activeUser}, onCall.ToSlice())

	// one more nanosecond: the active override has ended
	onCall = info.calcLatestOnCall(now.Add(time.Minute + time.Nanosecond))
	require.ElementsMatch(t, []uuid.UUID{futureUser}, onCall.ToSlice())
}
