-- name: IntKeyGetServiceID :one
SELECT
    service_id
FROM
    integration_keys
WHERE
    id = $1
    AND type = $2;

-- name: IntKeyCreate :exec
INSERT INTO integration_keys(id, name, type, service_id, external_system_name)
    VALUES ($1, $2, $3, $4, $5);

-- name: IntKeyFindOne :one
SELECT
    id,
    name,
    type,
    service_id,
    external_system_name
FROM
    integration_keys
WHERE
    id = $1;

-- name: IntKeyFindByService :many
SELECT
    id,
    name,
    type,
    service_id,
    external_system_name
FROM
    integration_keys
WHERE
    service_id = $1;

-- name: IntKeyDelete :exec
DELETE FROM integration_keys
WHERE id = ANY (@ids::uuid[]);

-- name: IntKeyGetConfig :one
SELECT
    config
FROM
    uik_config
WHERE
    id = $1
FOR UPDATE;

-- name: IntKeySetConfig :exec
INSERT INTO uik_config(id, config)
    VALUES ($1, $2)
ON CONFLICT (id)
    DO UPDATE SET
        config = $2;

-- name: IntKeyDeleteConfig :exec
DELETE FROM uik_config
WHERE id = $1;

-- name: IntKeyGetType :one
SELECT
    type
FROM
    integration_keys
WHERE
    id = $1;

-- name: IntKeyPromoteSecondary :one
UPDATE
    uik_config
SET
    primary_token = secondary_token,
    primary_token_hint = secondary_token_hint,
    secondary_token = NULL,
    secondary_token_hint = NULL
WHERE
    id = $1
RETURNING
    primary_token_hint;

-- name: IntKeyTokenHints :one
SELECT
    primary_token_hint,
    secondary_token_hint
FROM
    uik_config
WHERE
    id = $1;

-- name: IntKeySetPrimaryToken :one
UPDATE
    uik_config
SET
    primary_token = $2,
    primary_token_hint = $3
WHERE
    id = $1
    AND primary_token IS NULL
RETURNING
    id;

-- name: IntKeySetSecondaryToken :one
UPDATE
    uik_config
SET
    secondary_token = $2,
    secondary_token_hint = $3
WHERE
    id = $1
    AND secondary_token IS NULL
    AND primary_token IS NOT NULL
RETURNING
    id;

-- name: IntKeyUIKValidateService :one
SELECT
    k.service_id
FROM
    uik_config c
    JOIN integration_keys k ON k.id = c.id
WHERE
    c.id = sqlc.arg(key_id)
    AND k.type = 'universal'
    AND (c.primary_token = sqlc.arg(token_id)
        OR c.secondary_token = sqlc.arg(token_id));

-- name: IntKeyDeleteSecondaryToken :exec
UPDATE
    uik_config
SET
    secondary_token = NULL,
    secondary_token_hint = NULL
WHERE
    id = $1;

-- name: IntKeyEnsureChannel :one
WITH insert_q AS (
INSERT INTO notification_channels(id, type, name, value, dest)
        VALUES ($1, 'DEST', '', '', $2)
    ON CONFLICT (dest)
        DO NOTHING
    RETURNING
        id)
    SELECT
        id
    FROM
        insert_q
    UNION
    SELECT
        id
    FROM
        notification_channels
    WHERE
        type = 'DEST'
            AND dest = $2;

-- name: IntKeyInsertSignalMessage :exec
INSERT INTO pending_signals(dest_id, service_id, params)
    VALUES ($1, $2, $3);

