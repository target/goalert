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

-- name: NotifChanDeleteMany :exec
DELETE FROM notification_channels
WHERE id = ANY ($1::uuid[]);

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

