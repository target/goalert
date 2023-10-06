-- name: IntKeyGetServiceID :one
SELECT
    service_id
FROM
    integration_keys
WHERE
    id = $1
    AND type = $2;

-- name: IntKeyCreate :exec
INSERT INTO integration_keys(id, name, type, service_id)
    VALUES ($1, $2, $3, $4);

-- name: IntKeyFindOne :one
SELECT
    id,
    name,
    type,
    service_id
FROM
    integration_keys
WHERE
    id = $1;

-- name: IntKeyFindByService :many
SELECT
    id,
    name,
    type,
    service_id
FROM
    integration_keys
WHERE
    service_id = $1;

-- name: IntKeyDelete :exec
DELETE FROM integration_keys
WHERE id = ANY (@ids::uuid[]);

-- name: IntKeyFindByServiceRule :many
SELECT
    i.id,
    i.name,
    i.type,
    i.service_id
FROM
    integration_keys i
    JOIN service_rule_integration_keys si ON si.integration_key_id = i.id
        AND si.service_rule_id = $1;

