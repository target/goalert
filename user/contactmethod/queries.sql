-- name: ContactMethodAdd :exec
INSERT INTO user_contact_methods(id, name, type, value, disabled, user_id, enable_status_updates)
    VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ContactMethodUpdate :exec
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

-- name: ContactMethodFineOne :one
SELECT
    *
FROM
    user_contact_methods
WHERE
    id = $1;

-- name: ContactMethodFindOneUpdate :one
SELECT
    *
FROM
    user_contact_methods
WHERE
    id = $1
FOR UPDATE;

-- name: ContactMethodFindMany :many
SELECT
    *
FROM
    user_contact_methods
WHERE
    id = ANY ($1::uuid[]);

-- name: ContactMethodFindAll :many
SELECT
    *
FROM
    user_contact_methods
WHERE
    user_id = $1;

-- name: ContactMethodLookupUserID :many
SELECT DISTINCT
    user_id
FROM
    user_contact_methods
WHERE
    id = ANY ($1::uuid[]);

-- name: ContactMethodEnableDisable :one
UPDATE
    user_contact_methods
SET
    disabled = $2
WHERE
    dest = $1
RETURNING
    id;

-- name: ContactMethodMetaTV :one
SELECT
    coalesce(metadata, '{}'),
    now()::timestamptz AS now
FROM
    user_contact_methods
WHERE
    type = $1
    AND value = $2;

-- name: ContactMethodUpdateMetaTV :exec
UPDATE
    user_contact_methods
SET
    metadata = jsonb_set(jsonb_set(metadata, '{CarrierV1}', @carrier_v1::jsonb), '{CarrierV1,UpdatedAt}',('"' || NOW()::timestamptz AT TIME ZONE 'UTC' || '"')::jsonb)
WHERE
    type = $1
    AND value = $2;

