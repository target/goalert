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

-- name: SignalSearch :many
SELECT
    id,
    service_rule_id,
    service_id,
    outgoing_payload,
    scheduled,
    timestamp
FROM
    signals
WHERE (id <> ALL (@omit::bigint[]))
/* array_length returns NULL for empty arrays*/
    AND (array_length(@any_service_id::uuid[], 1) IS NULL
        OR service_id = ANY (@any_service_id::uuid[]))
    AND (array_length(@any_service_rule_id::uuid[], 1) IS NULL
        OR service_rule_id = ANY (@any_service_rule_id::uuid[]))
    AND (sqlc.narg(before_time)::timestamptz IS NULL
        OR timestamp < @before_time)
    AND (sqlc.narg(not_before_time)::timestamptz IS NULL
        OR timestamp >= @not_before_time)
    AND (sqlc.narg(after_id)::bigint IS NULL
        OR (@sort_mode::int = 0
            AND (timestamp < @after_timestamp
                OR (timestamp = @after_timestamp
                    AND id < @after_id)))
        OR (@sort_mode = 1
            AND (timestamp > @after_timestamp
                OR (timestamp = @after_timestamp
                    AND id > @after_id))))
ORDER BY
    CASE @sort_mode::int
    WHEN 0 THEN
        ROW (timestamp,
            id)
    END DESC,
    CASE @sort_mode
    WHEN 1 THEN
        ROW (timestamp,
            id)
    END
LIMIT $1;

