-- name: SignalMgrGetPending :many
SELECT
    id,
    dest_id,
    service_id
FROM
    pending_signals
WHERE
    message_id IS NULL
FOR UPDATE
    SKIP LOCKED
LIMIT 100;

-- name: SignalMgrInsertMessage :exec
INSERT INTO outgoing_messages(id, message_type, service_id, channel_id)
    VALUES ($1, 'signal_message', $2, $3);

-- name: SignalMgrUpdateSignal :exec
UPDATE
    pending_signals
SET
    message_id = $2
WHERE
    id = $1;

-- name: SignalMgrDeleteStale :exec
DELETE FROM pending_signals
WHERE message_id IS NULL
    AND created_at < NOW() - INTERVAL '1 hour';

