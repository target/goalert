-- name: SchedMgrRules :many
SELECT
    rule.*,
    coalesce(rule.tgt_user_id, part.user_id) AS resolved_user_id
FROM
    schedule_rules rule
    LEFT JOIN rotation_state rState ON rState.rotation_id = rule.tgt_rotation_id
    LEFT JOIN rotation_participants part ON part.id = rState.rotation_participant_id
WHERE
    coalesce(rule.tgt_user_id, part.user_id)
    NOTNULL;

-- name: SchedMgrRulesForSchedule :many
-- Returns the rules for a single schedule. The column list is identical to
-- SchedMgrRules so both rows convert to the same Go types.
SELECT
    rule.*,
    coalesce(rule.tgt_user_id, part.user_id) AS resolved_user_id
FROM
    schedule_rules rule
    LEFT JOIN rotation_state rState ON rState.rotation_id = rule.tgt_rotation_id
    LEFT JOIN rotation_participants part ON part.id = rState.rotation_participant_id
WHERE
    coalesce(rule.tgt_user_id, part.user_id)
    NOTNULL
    AND rule.schedule_id = @schedule_id;

-- name: SchedMgrOverrides :many
-- Returns the overrides that are active right now, for all schedules.
SELECT
    add_user_id,
    remove_user_id,
    tgt_schedule_id,
    start_time,
    end_time
FROM
    user_overrides
WHERE
    now() BETWEEN start_time AND end_time;

-- name: SchedMgrOverridesForSchedule :many
-- Returns every override for a single schedule that has not yet ended,
-- including ones that have not started; whether an override is currently
-- active is decided in Go (see overrideActive).
SELECT
    add_user_id,
    remove_user_id,
    tgt_schedule_id,
    start_time,
    end_time
FROM
    user_overrides
WHERE
    tgt_schedule_id = @schedule_id
    AND end_time > now();

-- name: SchedMgrDataForUpdate :many
SELECT
    schedule_id,
    data
FROM
    schedule_data
WHERE
    data NOTNULL
FOR UPDATE;

-- name: SchedMgrDataForSchedule :one
-- Returns the data for a single schedule, locking the row for the rest of the
-- transaction. The lock is required because schedule/storedata.go locks only
-- schedule_data (not schedules) when users edit, so this is the row that
-- serializes the engine against user edits of the same schedule.
SELECT
    data
FROM
    schedule_data
WHERE
    schedule_id = @schedule_id
FOR UPDATE;

-- name: SchedMgrSetData :exec
UPDATE
    schedule_data
SET
    data = $2
WHERE
    schedule_id = $1;

-- name: SchedMgrTimezones :many
SELECT
    id,
    time_zone
FROM
    schedules;

-- name: SchedMgrLockSchedule :one
-- Locks the schedule row for the rest of the transaction and returns its time
-- zone. FOR NO KEY UPDATE serializes concurrent jobs for the same schedule
-- without blocking the FOR KEY SHARE locks that foreign-key checks take when
-- rules or overrides referencing the schedule are inserted. sql.ErrNoRows
-- means the schedule was deleted.
SELECT
    time_zone
FROM
    schedules
WHERE
    id = @schedule_id
FOR NO KEY UPDATE;

-- name: SchedMgrOnCall :many
SELECT
    schedule_id,
    user_id
FROM
    schedule_on_call_users
WHERE
    end_time ISNULL;

-- name: SchedMgrOnCallForSchedule :many
-- Returns the users currently recorded as on-call for a single schedule.
SELECT
    user_id
FROM
    schedule_on_call_users
WHERE
    schedule_id = @schedule_id
    AND end_time ISNULL;

-- name: SchedMgrVerifyUsers :many
-- Returns the subset of the given user IDs that exist in the users table.
SELECT
    id
FROM
    users
WHERE
    id = ANY (sqlc.arg(user_ids)::uuid[]);

-- name: SchedMgrStartOnCall :exec
INSERT INTO schedule_on_call_users(schedule_id, start_time, user_id)
SELECT
    $1,
    now(),
    $2
FROM
    users
WHERE
    id = $2;

-- name: SchedMgrEndOnCall :exec
UPDATE
    schedule_on_call_users
SET
    end_time = now()
WHERE
    schedule_id = $1
    AND user_id = $2
    AND end_time ISNULL;

-- name: SchedMgrInsertMessage :exec
INSERT INTO outgoing_messages(id, message_type, channel_id, schedule_id)
    VALUES ($1, 'schedule_on_call_notification', $2, $3);
