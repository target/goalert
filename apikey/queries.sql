-- name: APIKeyInsert :exec
INSERT INTO gql_api_keys(id, name, description, POLICY, created_by, updated_by, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: APIKeyUpdate :exec
UPDATE
    gql_api_keys
SET
    name = $2,
    description = $3,
    updated_by = $4
WHERE
    id = $1;

-- name: APIKeyForUpdate :one
SELECT
    name,
    description
FROM
    gql_api_keys
WHERE
    id = $1
    AND deleted_at IS NULL
FOR UPDATE;

-- name: APIKeyDelete :exec
DELETE FROM gql_api_keys
WHERE id = $1;

-- name: APIKeyRecordUsage :exec
-- APIKeyRecordUsage records the usage of an API key.
INSERT INTO gql_api_key_usage(api_key_id, user_agent, ip_address)
    VALUES (@key_id::uuid, @user_agent::text, @ip_address::inet);

-- name: APIKeyAuthPolicy :one
-- APIKeyAuth returns the API key policy with the given id, if it exists and is not expired.
SELECT
    gql_api_keys.policy
FROM
    gql_api_keys
WHERE
    gql_api_keys.id = $1
    AND gql_api_keys.deleted_at IS NULL
    AND gql_api_keys.expires_at > now();

-- name: APIKeyList :many
-- APIKeyList returns all API keys, along with the last time they were used.
SELECT DISTINCT ON (gql_api_keys.id)
    gql_api_keys.*,
    gql_api_key_usage.used_at AS last_used_at,
    gql_api_key_usage.user_agent AS last_user_agent,
    gql_api_key_usage.ip_address AS last_ip_address
FROM
    gql_api_keys
    LEFT JOIN gql_api_key_usage ON gql_api_keys.id = gql_api_key_usage.api_key_id
WHERE
    gql_api_keys.deleted_at IS NULL
ORDER BY
    gql_api_keys.id,
    gql_api_key_usage.used_at DESC;

