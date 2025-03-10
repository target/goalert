-- name: AllPendingMsgDests :many
SELECT DISTINCT
    usr.name AS user_name,
    cm.dest AS cm_dest,
    nc.name AS nc_name,
    nc.dest AS nc_dest
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

-- name: ServiceAlertStats :many
-- ServiceAlertStats returns statistics about alerts for a service.
SELECT
    date_bin(sqlc.arg(stride)::interval, closed_at, sqlc.arg(origin)::timestamptz)::timestamptz AS bucket,
    coalesce(EXTRACT(EPOCH FROM AVG(time_to_ack)), 0)::double precision AS avg_time_to_ack_seconds,
    coalesce(EXTRACT(EPOCH FROM AVG(time_to_close)), 0)::double precision AS avg_time_to_close_seconds,
    coalesce(COUNT(*), 0)::bigint AS alert_count,
    coalesce(SUM(
            CASE WHEN escalated THEN
                1
            ELSE
                0
            END), 0)::bigint AS escalated_count
FROM
    alert_metrics
WHERE
    service_id = $1
    AND (closed_at BETWEEN sqlc.arg(start_time)
        AND sqlc.arg(end_time))
GROUP BY
    bucket
ORDER BY
    bucket;

