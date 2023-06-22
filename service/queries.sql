-- name: ServiceFindMany :many
SELECT
    s.id,
    s.name,
    s.description,
    s.escalation_policy_id,
    fav IS DISTINCT FROM NULL AS is_user_favorite,
    s.maintenance_expires_at
FROM
    services s
    LEFT JOIN user_favorites fav ON s.id = fav.tgt_service_id
        AND fav.user_id = @user_id
WHERE
    s.id = ANY (@service_ids::uuid[]);

-- name: ServiceFindOneForUpdate :one
SELECT
    s.id,
    s.name,
    s.description,
    s.escalation_policy_id,
    s.maintenance_expires_at
FROM
    services s
WHERE
    s.id = $1;

-- name: ServiceFindManyByEP :many
SELECT
    s.id,
    s.name,
    s.description,
    s.escalation_policy_id,
    fav IS DISTINCT FROM NULL AS is_user_favorite,
    s.maintenance_expires_at
FROM
    services s
    LEFT JOIN user_favorites fav ON s.id = fav.tgt_service_id
        AND fav.user_id = $2
WHERE
    s.escalation_policy_id = $1;

-- name: ServiceInsert :exec
INSERT INTO services(id, name, description, escalation_policy_id)
    VALUES ($1, $2, $3, $4);

-- name: ServiceUpdate :exec
UPDATE
    services
SET
    name = $2,
    description = $3,
    escalation_policy_id = $4,
    maintenance_expires_at = $5
WHERE
    id = $1;

-- name: ServiceDeleteMany :exec
DELETE FROM services
WHERE id = ANY (@service_ids::uuid[]);

