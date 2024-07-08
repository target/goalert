-- name: SchedMgrGetNCIDs :many
SELECT
    id
FROM
    notification_channels;

-- name: SchedMgrNCDedupMapping :many
SELECT
    old_id,
    new_id
FROM
    notification_channel_duplicates;

-- name: SchedMgrDataIDs :many
SELECT
    schedule_id
FROM
    schedule_data;

-- name: SchedMgrGetData :one
SELECT
    data
FROM
    schedule_data
WHERE
    schedule_id = $1;

-- name: SchedMgrSetDataV1Rules :exec
UPDATE
    schedule_data
SET
    data = jsonb_set(data, '{V1,OnCallNotificationRules}', $2)
WHERE
    schedule_id = $1;

