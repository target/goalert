-- name: EPStepActionsByStepId :many
SELECT
    a.user_id,
    a.schedule_id,
    a.rotation_id,
    ch.dest
FROM
    escalation_policy_actions a
    LEFT JOIN notification_channels ch ON a.channel_id = ch.id
WHERE
    a.escalation_policy_step_id = $1;

-- name: EPStepActionsAddAction :exec
INSERT INTO escalation_policy_actions(escalation_policy_step_id, user_id, schedule_id, rotation_id, channel_id)
    VALUES ($1, $2, $3, $4, $5);

-- name: EPStepActionsDeleteAction :exec
DELETE FROM escalation_policy_actions
WHERE escalation_policy_step_id = $1
    AND (user_id = $2
        OR schedule_id = $3
        OR rotation_id = $4
        OR channel_id = $5);

