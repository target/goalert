-- name: LogEvents :many
SELECT id,
    TIMESTAMP,
    DATA
FROM switchover_log
WHERE id > $1
ORDER BY id ASC
LIMIT 100;

-- name: LastLogID :one
SELECT COALESCE(MAX(id), 0)::bigint
FROM switchover_log;
