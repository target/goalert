-- name: MessageMgrGetPending :many
SELECT
    msg.id,
    msg.message_type,
    cm.id AS cm_id,
    cm.dest AS cm_dest,
    chan.id AS chan_id,
    chan.dest AS chan_dest,
    msg.alert_id,
    msg.alert_log_id,
    msg.user_verification_code_id,
    cm.user_id,
    msg.service_id,
    msg.created_at,
    msg.sent_at,
    msg.status_alert_ids,
    msg.schedule_id
FROM
    outgoing_messages msg
    LEFT JOIN user_contact_methods cm ON cm.id = msg.contact_method_id
    LEFT JOIN notification_channels chan ON chan.id = msg.channel_id
WHERE
    sent_at >= $1
    OR last_status = 'pending'
    AND (msg.contact_method_id ISNULL
        OR msg.message_type = 'verification_message'
        OR NOT cm.disabled);

