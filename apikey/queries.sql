-- name: APIKeyInsert :exec
INSERT INTO gql_api_keys(id, name, description, POLICY, created_by, updated_by, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: APIKeyDelete :exec
DELETE FROM gql_api_keys
WHERE id = $1;

-- name: APIKeyRecordUsage :exec
-- APIKeyRecordUsage records the usage of an API key.
INSERT INTO gql_api_key_usage(api_key_id, user_agent, ip_address)
    VALUES (@key_id::uuid, @user_agent::text, @ip_address::inet)
ON CONFLICT (api_key_id)
    DO UPDATE SET
        used_at = now(), user_agent = @user_agent::text, ip_address = @ip_address::inet;

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

-- name: APIKeyAuthCheck :one
SELECT
    TRUE
FROM
    gql_api_keys
WHERE
    gql_api_keys.id = $1
    AND gql_api_keys.deleted_at IS NULL
    AND gql_api_keys.expires_at > now();

