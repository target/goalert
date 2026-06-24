-- name: EngineGetSignalParams :one
-- Get a pending signal's rendered params.
SELECT
    params
FROM
    pending_signals
WHERE
    message_id = $1;

-- name: EngineIsKnownDest :one
-- Check if a destination is known in user_contact_methods or notification_channels table.
SELECT
    EXISTS (
        SELECT
        FROM
            user_contact_methods uc
        WHERE
            uc.dest = $1
            AND uc.disabled = FALSE)
    OR EXISTS (
        SELECT
        FROM
            notification_channels nc
        WHERE
            nc.dest = $1);

