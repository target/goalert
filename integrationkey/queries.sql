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
    id = $1;

-- name: IntKeySetConfig :exec
INSERT INTO uik_config(id, config)
    VALUES ($1, $2)
ON CONFLICT (id)
    DO UPDATE SET
        config = $2;

-- name: IntKeyDeleteConfig :exec
DELETE FROM uik_config
WHERE id = $1;

