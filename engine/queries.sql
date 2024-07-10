-- name: EngineGetSignalParams :one
-- Get a pending signal's rendered params.
SELECT
    params
FROM
    pending_signals
WHERE
    message_id = $1;

