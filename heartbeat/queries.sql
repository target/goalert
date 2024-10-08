-- name: HBInsert :exec
-- Inserts a new heartbeat record
INSERT INTO heartbeat_monitors(id, name, service_id, heartbeat_interval, additional_details, disable_reason)
    VALUES (@id, @name, @service_id, @heartbeat_interval, @additional_details, @disable_reason);

-- name: HBByService :many
-- Returns all heartbeat records for a service
SELECT
    *
FROM
    heartbeat_monitors
WHERE
    service_id = @service_id;

-- name: HBManyByID :many
SELECT
    *
FROM
    heartbeat_monitors
WHERE
    id = ANY (@ids::uuid[]);

-- name: HBDelete :exec
DELETE FROM heartbeat_monitors
WHERE id = ANY (@id::uuid[]);

-- name: HBUpdate :exec
UPDATE
    heartbeat_monitors
SET
    name = @name,
    heartbeat_interval = @heartbeat_interval,
    additional_details = @additional_details,
    disable_reason = @disable_reason
WHERE
    id = @id;

-- name: HBByIDForUpdate :one
SELECT
    *
FROM
    heartbeat_monitors
WHERE
    id = @id
FOR UPDATE;

-- name: HBRecordHeartbeat :exec
UPDATE
    heartbeat_monitors
SET
    last_heartbeat = now()
WHERE
    id = @id;

