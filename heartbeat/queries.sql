-- name: HBInsert :exec
-- HBInsert will insert a new heartbeat record.
INSERT INTO heartbeat_monitors(id, name, service_id, heartbeat_interval, additional_details, muted)
    VALUES (@id, @name, @service_id, @heartbeat_interval, @additional_details, @muted);

-- name: HBByService :many
-- HBByService returns all heartbeat records for a service.
SELECT
    *
FROM
    heartbeat_monitors
WHERE
    service_id = @service_id;

-- name: HBManyByID :many
-- HBManyByID returns multiple heartbeat records by their IDs.
SELECT
    *
FROM
    heartbeat_monitors
WHERE
    id = ANY (@ids::uuid[]);

-- name: HBDelete :exec
-- HBDelete will delete a heartbeat record.
DELETE FROM heartbeat_monitors
WHERE id = ANY (@id::uuid[]);

-- name: HBUpdate :exec
-- HBUpdate will update a heartbeat record.
UPDATE
    heartbeat_monitors
SET
    name = @name,
    heartbeat_interval = @heartbeat_interval,
    additional_details = @additional_details,
    muted = @muted
WHERE
    id = @id;

-- name: HBByIDForUpdate :one
-- HBByIDForUpdate returns a single heartbeat record by ID for update.
SELECT
    *
FROM
    heartbeat_monitors
WHERE
    id = @id
FOR UPDATE;

-- name: HBRecordHeartbeat :exec
-- HBRecordHeartbeat updates the last heartbeat time for a monitor.
UPDATE
    heartbeat_monitors
SET
    last_heartbeat = now()
WHERE
    id = @id;

