-- name: EnableChangeLogTriggers :exec
UPDATE
    switchover_state
SET
    current_state = 'in_progress'
WHERE
    current_state = 'idle';

-- name: DisableChangeLogTriggers :exec
UPDATE
    switchover_state
SET
    current_state = 'idle'
WHERE
    current_state = 'in_progress';

-- name: Now :one
SELECT
    now()::timestamptz;

-- name: ActiveTxCount :one
SELECT
    COUNT(*)
FROM
    pg_stat_activity
WHERE
    "state" <> 'idle'
    AND "xact_start" <= $1;

