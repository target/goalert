-- name: SWOConnLock :one
WITH LOCK AS (
    SELECT
        pg_advisory_lock_shared(4369))
SELECT
    current_state = 'use_next_db'
FROM
    LOCK,
    switchover_state;

-- name: SWOConnUnlockAll :exec
SELECT
    pg_advisory_unlock_all();

