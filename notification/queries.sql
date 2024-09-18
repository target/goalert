-- name: NfyLastMessageStatus :one
SELECT
    sqlc.embed(om),
    cm.dest AS cm_dest,
    ch.dest AS ch_dest
FROM
    outgoing_messages om
    LEFT JOIN notification_channels ch ON om.channel_id = ch.id
    LEFT JOIN user_contact_methods cm ON om.contact_method_id = cm.id
WHERE
    message_type = $1
    AND contact_method_id = $2
    AND om.created_at >= $3;

-- name: NfyManyMessageStatus :many
SELECT
    sqlc.embed(om),
    cm.dest AS cm_dest,
    ch.dest AS ch_dest
FROM
    outgoing_messages om
    LEFT JOIN notification_channels ch ON om.channel_id = ch.id
    LEFT JOIN user_contact_methods cm ON om.contact_method_id = cm.id
WHERE
    om.id = ANY ($1::uuid[]);

-- name: NfyOriginalMessageStatus :one
SELECT
    sqlc.embed(om),
    cm.dest AS cm_dest,
    ch.dest AS ch_dest
FROM
    outgoing_messages om
    LEFT JOIN notification_channels ch ON om.channel_id = ch.id
    LEFT JOIN user_contact_methods cm ON om.contact_method_id = cm.id
WHERE
    message_type = 'alert_notification'
    AND alert_id = $1
    AND (contact_method_id = $2
        OR channel_id = $3)
ORDER BY
    sent_at
LIMIT 1;

