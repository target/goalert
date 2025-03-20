-- name: AlertLog_InsertEP :exec
-- Inserts a new alert log for all alerts in the escalation policy that are not closed.
INSERT INTO alert_logs(alert_id, event, sub_type, sub_user_id, sub_integration_key_id, sub_hb_monitor_id, sub_channel_id, sub_classifier, meta, message)
SELECT
    a.id,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
FROM
    alerts a
    JOIN services svc ON svc.id = a.service_id
        AND svc.escalation_policy_id = $1
WHERE
    a.status != 'closed';

-- name: AlertLog_InsertSvc :exec
-- Inserts a new alert log for all alerts in the service that are not closed.
INSERT INTO alert_logs(alert_id, event, sub_type, sub_user_id, sub_integration_key_id, sub_hb_monitor_id, sub_channel_id, sub_classifier, meta, message)
SELECT
    a.id,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
FROM
    alerts a
WHERE
    a.service_id = $1
    AND (($2 = 'closed'::enum_alert_log_event
            AND a.status != 'closed')
        OR ($2::enum_alert_log_event IN ('acknowledged', 'notification_sent')
            AND a.status = 'triggered'));

-- name: AlertLog_InsertMany :exec
-- Inserts many alert logs
INSERT INTO alert_logs(alert_id, event, sub_type, sub_user_id, sub_integration_key_id, sub_hb_monitor_id, sub_channel_id, sub_classifier, meta, message)
SELECT
    unnest,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
FROM
    unnest($1::bigint[]);

-- name: AlertLog_LookupCMDest :one
-- Looks up the destination for a contact method
SELECT
    dest
FROM
    user_contact_methods
WHERE
    id = $1;

-- name: AlertLog_HBIntervalMinutes :one
-- Looks up the heartbeat interval in minutes for a heartbeat monitor
SELECT
    (EXTRACT(EPOCH FROM heartbeat_interval) / 60)::int
FROM
    heartbeat_monitors
WHERE
    id = $1;

-- name: AlertLog_LookupCallbackDest :one
-- Looks up the destination for a callback
SELECT
    coalesce(cm.dest, ch.dest) AS dest
FROM
    outgoing_messages log
    LEFT JOIN user_contact_methods cm ON cm.id = log.contact_method_id
    LEFT JOIN notification_channels ch ON ch.id = log.channel_id
WHERE
    log.id = $1;

-- name: AlertLog_LookupNCDest :one
SELECT
    dest
FROM
    notification_channels
WHERE
    id = $1;

