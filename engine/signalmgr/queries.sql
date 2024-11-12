-- name: SignalMgrGetPending :many
-- Get a batch of pending signals to process.
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
-- Insert a new message into the outgoing_messages table.
INSERT INTO outgoing_messages(id, message_type, service_id, channel_id)
    VALUES ($1, 'signal_message', $2, $3);

-- name: SignalMgrUpdateSignal :exec
-- Update a pending signal with the message_id.
UPDATE
    pending_signals
SET
    message_id = $2
WHERE
    id = $1;

-- name: SignalMgrDeleteStale :exec
-- Delete stale pending signals.
DELETE FROM pending_signals
WHERE message_id IS NULL
    AND created_at < NOW() - INTERVAL '1 hour';

-- name: SignalMgrGetScheduled :many
SELECT
    count(*),
    service_id,
    channel_id
FROM
    outgoing_messages
WHERE
    message_type = 'signal_message'
    AND last_status = 'pending'
GROUP BY
    service_id,
    channel_id;

