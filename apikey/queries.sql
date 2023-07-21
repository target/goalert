-- name: APIKeyInsert :exec
INSERT INTO api_keys(id, version, user_id, service_id, name, data, expires_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: APIKeyDelete :exec
DELETE FROM api_keys
WHERE id = $1;

-- name: APIKeyGet :one
SELECT
    *
FROM
    api_keys
WHERE
    id = $1;

-- name: APIKeyAuth :one
UPDATE
    api_keys
SET
    last_used_at = now()
WHERE
    id = $1
RETURNING
    *;

