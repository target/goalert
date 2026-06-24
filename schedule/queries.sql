-- name: ScheduleFindManyByUser :many
SELECT
    *
FROM
    schedules
WHERE
    id = ANY (
        SELECT
            schedule_id
        FROM
            schedule_rules
        WHERE
            tgt_user_id = $1
            OR tgt_rotation_id = ANY (
                SELECT
                    rotation_id
                FROM
                    rotation_participants
                WHERE
                    user_id = $1));

-- name: SchedFindData :one
-- Returns the schedule data for a given schedule ID.
SELECT
    data
FROM
    schedule_data
WHERE
    schedule_id = $1;

-- name: SchedFindDataForUpdate :one
-- Returns the schedule data for a given schedule ID with FOR UPDATE lock.
SELECT
    data
FROM
    schedule_data
WHERE
    schedule_id = $1
FOR UPDATE;

-- name: SchedInsertData :exec
-- Inserts empty schedule data for a given schedule ID.
INSERT INTO schedule_data (schedule_id, data)
VALUES ($1, '{}');

-- name: SchedUpdateData :exec
-- Updates the schedule data for a given schedule ID.
UPDATE schedule_data
SET data = $2
WHERE schedule_id = $1;

-- name: SchedCreate :one
-- Creates a new schedule and returns its ID.
INSERT INTO schedules (id, name, description, time_zone)
VALUES (DEFAULT, $1, $2, $3)
RETURNING id;

-- name: SchedUpdate :exec
-- Updates an existing schedule.
UPDATE schedules
SET name = $2, description = $3, time_zone = $4
WHERE id = $1;

-- name: SchedFindAll :many
-- Returns all schedules.
SELECT id, name, description, time_zone
FROM schedules;

-- name: SchedFindOne :one
-- Returns a single schedule with user favorite status.
SELECT
    s.id,
    s.name,
    s.description,
    s.time_zone,
    fav IS DISTINCT FROM NULL as is_favorite
FROM schedules s
LEFT JOIN user_favorites fav ON
    fav.tgt_schedule_id = s.id AND fav.user_id = $2
WHERE s.id = $1;

-- name: SchedFindOneForUpdate :one
-- Returns a single schedule with FOR UPDATE lock.
SELECT id, name, description, time_zone
FROM schedules
WHERE id = $1
FOR UPDATE;

-- name: SchedFindMany :many
-- Returns multiple schedules with user favorite status.
SELECT
    s.id,
    s.name,
    s.description,
    s.time_zone,
    fav IS DISTINCT FROM NULL as is_favorite
FROM schedules s
LEFT JOIN user_favorites fav ON
    fav.tgt_schedule_id = s.id AND fav.user_id = $2
WHERE s.id = ANY($1::uuid[]);

-- name: SchedDeleteMany :exec
-- Deletes multiple schedules by their IDs.
DELETE FROM schedules
WHERE id = ANY($1::uuid[]);

