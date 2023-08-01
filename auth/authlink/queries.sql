-- name: AuthLinkAddReq :exec
INSERT INTO auth_link_requests(id, provider_id, subject_id, expires_at, metadata)
    VALUES ($1, $2, $3, $4, $5);

-- name: AuthLinkUseReq :one
DELETE FROM auth_link_requests
WHERE id = $1
    AND expires_at > now()
RETURNING
    provider_id,
    subject_id;

-- name: AuthLinkAddAuthSubject :exec
INSERT INTO auth_subjects(provider_id, subject_id, user_id)
    VALUES ($1, $2, $3);

-- name: AuthLinkMetadata :one
SELECT
    metadata
FROM
    auth_link_requests
WHERE
    id = $1
    AND expires_at > now();

