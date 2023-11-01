-- name: OutgoingSignalFindNext :one
SELECT
    id,
    signal_id,
    service_id,
    sent_at,
    destination_type,
    destination_id,
    destination_val,
    content
FROM
    outgoing_signals
WHERE
    sent_at IS NULL
LIMIT 1
FOR UPDATE
    SKIP LOCKED;

-- name: OutgoingSignalUpdateSent :exec
UPDATE
    outgoing_signals
SET
    sent_at = now()
WHERE
    id = $1
    AND sent_at IS NULL;

