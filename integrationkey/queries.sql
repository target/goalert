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

