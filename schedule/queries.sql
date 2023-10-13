SELECT
    *
FROM
    schedules
WHERE
    id = ANY (
        SELECT
            schedule_id
        FROM
            schedule_rules
        WHERE
            tgt_user_id = $1
            OR tgt_rotation_id = ANY (
                SELECT
                    rotation_id
                FROM
                    rotation_participants
                WHERE
                    user_id = $1));