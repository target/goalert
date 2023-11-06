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
    tgt_service_id = $1;

-- name: LabelDeleteKeyByTarget :exec
DELETE FROM labels
WHERE key = $1
    AND tgt_service_id = $2;

-- name: LabelSetByTarget :exec
INSERT INTO labels(key, value, tgt_service_id)
    VALUES ($1, $2, $3)
ON CONFLICT (key, tgt_service_id)
    DO UPDATE SET
        value = $2;

