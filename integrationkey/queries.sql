-- name: GetServiceID :one
SELECT service_id FROM integration_keys
WHERE id = $1 AND type = $2;

-- name: CreateIntegrationKey :exec
INSERT INTO integration_keys (id, name, type, service_id)
VALUES ($1, $2, $3, $4);

-- name: FindOneIntegrationKey :one
SELECT id, name, type, service_id 
FROM integration_keys
WHERE id = $1;

-- name: FindIntegrationKeysByService :many
SELECT id, name, type, service_id 
FROM integration_keys
WHERE service_id = $1;

-- name: DeleteIntegrationKey :exec
DELETE FROM integration_keys WHERE id = any(@ids::uuid[]);
