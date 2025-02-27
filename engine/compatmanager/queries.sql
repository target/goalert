-- name: CompatAuthSubSlackMissingCM :many
-- Get up to 10 auth_subjects (slack only) missing a contact method.
SELECT
    *
FROM
    auth_subjects
WHERE
    provider_id LIKE 'slack:%'
    AND cm_id IS NULL
FOR UPDATE
    SKIP LOCKED
LIMIT 10;

-- name: CompatAuthSubSetCMID :exec
-- Updates the contact method id for an auth_subject with the given destination.
UPDATE
    auth_subjects
SET
    cm_id =(
        SELECT
            id
        FROM
            user_contact_methods
        WHERE
            type = 'SLACK_DM'
            AND value = $2)
WHERE
    auth_subjects.id = $1;

-- name: CompatInsertUserCM :exec
-- Inserts a new contact method for a user.
INSERT INTO user_contact_methods(id, name, type, value, user_id, pending)
    VALUES ($1, $2, $3, $4, $5, FALSE)
ON CONFLICT (type, value)
    DO NOTHING;

-- name: CompatCMMissingSub :many
-- Get up to 10 contact methods missing an auth_subjects link.
SELECT
    id,
    user_id,
    value
FROM
    user_contact_methods
WHERE
    type = 'SLACK_DM'
    AND NOT disabled
    AND NOT EXISTS (
        SELECT
            1
        FROM
            auth_subjects
        WHERE
            cm_id = user_contact_methods.id)
FOR UPDATE
    SKIP LOCKED
LIMIT 10;

-- name: CompatUpsertAuthSubject :exec
-- Inserts a new auth_subject for a user.
INSERT INTO auth_subjects(user_id, subject_id, provider_id, cm_id)
    VALUES ($1, $2, $3, $4)
ON CONFLICT (subject_id, provider_id)
    DO UPDATE SET
        user_id = $1, cm_id = $4;

