-- name: SchedMgrNCDedupMapping :many
-- Returns the mapping of old notification channel IDs to new notification channel IDs.
SELECT
    old_id,
    new_id
FROM
    notification_channel_duplicates;

-- name: SchedMgrDataIDs :many
-- Returns all schedule IDs that have an entry in the schedule_data table.
SELECT
    schedule_id
FROM
    schedule_data;

-- name: SchedMgrGetData :one
-- Returns the data for a single schedule.
SELECT
    data
FROM
    schedule_data
WHERE
    schedule_id = $1;

-- name: SchedMgrSetDataV1Rules :exec
-- Sets the .V1.OnCallNotificationRules for a schedule.
UPDATE
    schedule_data
SET
    data = jsonb_set(data, '{V1,OnCallNotificationRules}', $2)
WHERE
    schedule_id = $1;

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

-- name: SchedMgrOverrides :many
SELECT
    add_user_id,
    remove_user_id,
    tgt_schedule_id
FROM
    user_overrides
WHERE
    now() BETWEEN start_time AND end_time;

-- name: SchedMgrDataForUpdate :many
SELECT
    schedule_id,
    data
FROM
    schedule_data
WHERE
    data NOTNULL
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

-- name: SchedMgrOnCall :many
SELECT
    schedule_id,
    user_id
FROM
    schedule_on_call_users
WHERE
    end_time ISNULL;

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

