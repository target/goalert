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

