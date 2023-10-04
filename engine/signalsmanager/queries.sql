-- name: SignalsManagerFindNext :one
SELECT
    id,
    service_rule_id,
    service_id,
    outgoing_payload,
    scheduled,
    timestamp
FROM
    signals
WHERE
    NOT scheduled
LIMIT 1
FOR UPDATE
    SKIP LOCKED;

-- name: SignalsManagerSetScheduled :exec
UPDATE
    signals
SET
    scheduled = NOT scheduled
WHERE
    NOT scheduled
    AND id = $1;

-- name: SignalsManagerSendOutgoing :exec
INSERT INTO outgoing_signals (id, service_id, outgoing_payload, channel_id)
    VALUES ($1, $2, $3, $4);

