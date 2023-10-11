-- name: FindManyByAssignments :many
SELECT * FROM schedules
WHERE id = any(SELECT schedule_id FROM schedule_rules WHERE tgt_user_id = $1 OR tgt_rotation_id = any(@TgtRotationIDs::UUID[]));