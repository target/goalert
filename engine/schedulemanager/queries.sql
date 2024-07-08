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

