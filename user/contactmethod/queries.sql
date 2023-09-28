-- name: AddContactMethod :exec
INSERT INTO user_contact_methods(id, name, type, value, disabled, user_id, enable_status_updates)
    VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: UpdateContactMethod :exec
UPDATE
    user_contact_methods
SET
    name = $2,
    disabled = $3,
    enable_status_updates = $4
WHERE
    id = $1;

-- name: DeleteContactMethod :exec
DELETE FROM user_contact_methods
WHERE id = ANY ($1::uuid[]);

-- name: FindOneContactMethod :one
SELECT
    id,
    name,
    type,
    value,
    disabled,
    user_id,
    last_test_verify_at,
    enable_status_updates,
    pending
FROM
    user_contact_methods
WHERE
    id = $1;

-- name: FindOneUpdateContactMethod :one
SELECT
    id,
    name,
    type,
    value,
    disabled,
    user_id,
    last_test_verify_at,
    enable_status_updates,
    pending
FROM
    user_contact_methods
WHERE
    id = $1
FOR UPDATE;

-- name: FindManyContactMethod :many
SELECT
    id,
    name,
    type,
    value,
    disabled,
    user_id,
    last_test_verify_at,
    enable_status_updates,
    pending
FROM
    user_contact_methods
WHERE
    id = ANY ($1::uuid[]);

-- name: FindAllContactMethod :many
SELECT
    id,
    name,
    type,
    value,
    disabled,
    user_id,
    last_test_verify_at,
    enable_status_updates,
    pending
FROM
    user_contact_methods
WHERE
    user_id = $1;

-- name: LookupUserIDContactMethod :many
SELECT DISTINCT
    user_id
FROM
    user_contact_methods
WHERE
    id = ANY ($1::uuid[]);

-- name: EnableContactMethod :one
UPDATE
    user_contact_methods
SET
    disabled = FALSE
WHERE
    type = $1
    AND value = $2
RETURNING
    id;

-- name: MetaTVContactMethod :one
SELECT
    coalesce(metadata, '{}'),
    now()::timestamptz AS now
FROM
    user_contact_methods
WHERE
    type = $1
    AND value = $2;

-- name: UpdateMetaTVContactMethod :exec
UPDATE
    user_contact_methods
SET
    metadata = jsonb_set(jsonb_set(metadata, '{CarrierV1}', @carrier_v1::jsonb), '{CarrierV1,UpdatedAt}',('"' || NOW()::timestamptz AT TIME ZONE 'UTC' || '"')::jsonb) 
WHERE
    type = $1
    AND value = $2;

-- name: DisableContactMethod :one
UPDATE
    user_contact_methods
SET
    disabled = TRUE
WHERE
    type = $1
    AND value = $2
RETURNING
    id;

