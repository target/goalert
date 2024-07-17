-- name: NotifChanFindOne :one
SELECT
    *
FROM
    notification_channels
WHERE
    id = $1;

-- name: NotifChanFindMany :many
SELECT
    *
FROM
    notification_channels
WHERE
    id = ANY ($1::uuid[]);

-- name: NotifChanCreate :exec
INSERT INTO notification_channels(id, name, type, value)
    VALUES ($1, $2, $3, $4);

-- name: NotifChanUpdateName :exec
UPDATE
    notification_channels
SET
    name = $2
WHERE
    id = $1;

-- name: NotifChanDeleteMany :exec
DELETE FROM notification_channels
WHERE id = ANY ($1::uuid[]);

-- name: NotifChanFindByValue :one
SELECT
    *
FROM
    notification_channels
WHERE
    type = $1
    AND value = $2;

-- name: NotifChanLock :exec
LOCK notification_channels IN SHARE ROW EXCLUSIVE MODE;

-- name: NotifChanUpsertDest :one
-- NotifChanUpsertDest will insert a new destination if it does not exist, or updating it's name if it does.
INSERT INTO notification_channels(id, dest, name)
    VALUES ($1, $2, $3)
ON CONFLICT (dest)
    DO UPDATE SET
        name = $3
    RETURNING
        id;

-- name: NotifChanFindDestID :one
SELECT
    id
FROM
    notification_channels
WHERE
    dest = $1;

