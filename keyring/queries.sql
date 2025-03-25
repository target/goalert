-- name: Keyring_LockKeyrings :exec
-- Locks the keyring table so no new keyrings can be created.
LOCK TABLE keyring IN ACCESS EXCLUSIVE MODE;

-- name: Keyring_LockConfig :exec
-- Locks the config table so no new config payloads can be created.
LOCK TABLE config IN ACCESS EXCLUSIVE MODE;

-- name: Keyring_GetKeyringSecrets :many
SELECT
    id,
    signing_key,
    next_key
FROM
    keyring;

-- name: Keyring_UpdateKeyringSecrets :exec
UPDATE
    keyring
SET
    signing_key = @signing_key,
    next_key = @next_key
WHERE
    id = @id;

-- name: Keyring_GetConfigPayloads :many
SELECT
    id,
    data
FROM
    config;

-- name: Keyring_UpdateConfigPayload :exec
UPDATE
    config
SET
    data = @data
WHERE
    id = @id;

