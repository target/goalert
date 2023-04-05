-- name: FindOneCalSub :one
SELECT id,
    NAME,
    user_id,
    disabled,
    schedule_id,
    config,
    last_access
FROM user_calendar_subscriptions
WHERE id = $1;

-- name: CalSubRenderInfo :one
SELECT
    now()::timestamptz AS now,
    sub.schedule_id,
    sched.name AS schedule_name,
    sub.config,
    sub.user_id
FROM
    user_calendar_subscriptions sub
    JOIN schedules sched ON sched.id = schedule_id
WHERE
    sub.id = $1;

-- name: FindOneCalSubForUpdate :one
SELECT id,
    NAME,
    user_id,
    disabled,
    schedule_id,
    config,
    last_access
FROM user_calendar_subscriptions
WHERE id = $1 FOR
UPDATE;

-- name: FindManyCalSubByUser :many
SELECT id,
    NAME,
    user_id,
    disabled,
    schedule_id,
    config,
    last_access
FROM user_calendar_subscriptions
WHERE user_id = $1;

-- name: DeleteManyCalSub :exec
DELETE FROM user_calendar_subscriptions
WHERE id = ANY($1::uuid [ ])
    AND user_id = $2;

-- name: CreateCalSub :one
INSERT INTO user_calendar_subscriptions (
        id,
        NAME,
        user_id,
        disabled,
        schedule_id,
        config
    )
VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at;

-- name: Now :one
SELECT now()::timestamptz;

-- name: CalSubAuthUser :one
UPDATE user_calendar_subscriptions
SET last_access = now()
WHERE NOT disabled
    AND id = $1
    AND date_trunc('second', created_at) = $2 RETURNING user_id;

-- name: UpdateCalSub :exec
UPDATE user_calendar_subscriptions
SET NAME = $1,
    disabled = $2,
    config = $3,
    last_update = now()
WHERE id = $4
    AND user_id = $5;
