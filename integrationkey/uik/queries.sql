-- name: UIKInsertSample :exec
INSERT INTO uik_samples(id, key_id, request_data, user_note, failed)
    VALUES ($1, $2, $3, $4, $5);

-- name: UIKGetSamples :many
SELECT
    *
FROM
    uik_samples
WHERE
    key_id = $1
ORDER BY
    created_at DESC;

-- name: UIKDeleteOldSamples :exec
DELETE FROM uik_samples
WHERE created_at < $1
    AND NOT failed
    AND user_note IS NULL;

