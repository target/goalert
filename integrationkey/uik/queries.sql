-- name: UIKInsertSample :exec
INSERT INTO uik_samples(id, key_id, request_data, user_note, failed)
    VALUES ($1, $2, $3, $4, $5);

-- name: UIKUpdateSample :exec
UPDATE
    uik_samples
SET
    user_note = $2,
    request_data = $3
WHERE
    id = $1;

-- name: UIKGetSamples :many
SELECT
    *
FROM
    uik_samples
WHERE
    key_id = $1
ORDER BY
    created_at DESC;

-- name: UIKDeleteSample :exec
DELETE FROM uik_samples
WHERE id = $1;

-- name: UIKDeleteOldestSamples :exec
DELETE FROM uik_samples
WHERE id = ANY (
        SELECT
            id
        FROM
            uik_samples s
        WHERE
            s.key_id = $1
            AND s.user_note IS NULL
            AND s.failed = $2
        ORDER BY
            created_at DESC OFFSET $3);

