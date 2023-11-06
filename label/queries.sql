-- name: LabelUniqueKeys :many
SELECT DISTINCT
    key
FROM
    labels;

-- name: LabelFindAllByTarget :many
SELECT
    key,
    value
FROM
    labels
WHERE
    tgt_service_id = sqlc.narg(service_id)::uuid
    OR tgt_schedule_id = sqlc.narg(schedule_id)::uuid
    OR tgt_rotation_id = sqlc.narg(rotation_id)::uuid
    OR tgt_ep_id = sqlc.narg(ep_id)::uuid;

-- name: LabelDeleteKeyByTarget :exec
DELETE FROM labels
WHERE key = $1
    AND (tgt_service_id = sqlc.narg(service_id)::uuid
        OR tgt_schedule_id = sqlc.narg(schedule_id)::uuid
        OR tgt_rotation_id = sqlc.narg(rotation_id)::uuid
        OR tgt_ep_id = sqlc.narg(ep_id)::uuid);

-- name: LabelSetByTarget :exec
INSERT INTO labels(key, value, tgt_service_id, tgt_schedule_id, tgt_rotation_id, tgt_ep_id)
    VALUES ($1, $2, sqlc.narg(service_id)::uuid, sqlc.narg(schedule_id)::uuid, sqlc.narg(rotation_id)::uuid, sqlc.narg(ep_id)::uuid)
ON CONFLICT (key, tgt_service_id, tgt_schedule_id, tgt_rotation_id, tgt_ep_id)
    DO UPDATE SET
        value = $2;

