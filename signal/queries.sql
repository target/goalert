-- name: SignalInsert :one
INSERT INTO signals(service_rule_id, service_id, outgoing_payload, scheduled)
    VALUES ($1, $2, $3, FALSE)
RETURNING
    id, timestamp;

-- name: SignalFindMany :many
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
    id = ANY (@ids::bigint[]);

