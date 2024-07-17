-- name: AllPendingMsgDests :many
SELECT DISTINCT
    usr.name AS user_name,
    cm.type AS cm_type,
    nc.name AS nc_name,
    nc.type AS nc_type
FROM
    outgoing_messages om
    LEFT JOIN users usr ON usr.id = om.user_id
    LEFT JOIN notification_channels nc ON nc.id = om.channel_id
    LEFT JOIN user_contact_methods cm ON cm.id = om.contact_method_id
WHERE
    om.last_status = 'pending'
    AND (now() - om.created_at) > INTERVAL '15 seconds'
    AND (om.alert_id = @alert_id::bigint
        OR (om.message_type = 'alert_notification_bundle'
            AND om.service_id = @service_id::uuid));

-- name: GQLUserOnCallOverview :many
SELECT
    svc.id AS service_id,
    svc.name AS service_name,
    ep.id AS policy_id,
    ep.name AS policy_name,
    step.step_number
FROM
    ep_step_on_call_users oc
    JOIN escalation_policy_steps step ON step.id = oc.ep_step_id
    JOIN escalation_policies ep ON ep.id = step.escalation_policy_id
    JOIN services svc ON svc.escalation_policy_id = ep.id
WHERE
    oc.user_id = $1;

-- name: GQLLookupNCDest :one
SELECT
    dest
FROM
    notification_channels
WHERE
    id = $1;

