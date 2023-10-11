-- name: FindManyByUserID :many
SELECT * FROM rotations
WHERE id = any(SELECT rotation_id FROM rotation_participants WHERE user_id = $1);