-- name: UserFavSet :exec
INSERT INTO user_favorites(user_id, tgt_service_id, tgt_schedule_id, tgt_rotation_id, tgt_escalation_policy_id, tgt_user_id)
    VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT
    DO NOTHING;

-- name: UserFavUnset :exec
DELETE FROM user_favorites
WHERE user_id = $1
    AND tgt_service_id = $2
    OR tgt_schedule_id = $3
    OR tgt_rotation_id = $4
    OR tgt_escalation_policy_id = $5
    OR tgt_user_id = $6;

-- name: UserFavFindAll :many
SELECT
    tgt_service_id,
    tgt_schedule_id,
    tgt_rotation_id,
    tgt_escalation_policy_id,
    tgt_user_id
FROM
    user_favorites
WHERE
    user_id = @user_id
    AND ((tgt_service_id NOTNULL
            AND @allow_services::bool)
        OR (tgt_schedule_id NOTNULL
            AND @allow_schedules::bool)
        OR (tgt_rotation_id NOTNULL
            AND @allow_rotations::bool)
        OR (tgt_escalation_policy_id NOTNULL
            AND @allow_escalation_policies::bool)
        OR (tgt_user_id NOTNULL
            AND @allow_users::bool));

