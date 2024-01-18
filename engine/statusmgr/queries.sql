-- name: StatusMgrUpdateCMForced :exec
UPDATE
    user_contact_methods
SET
    enable_status_updates = TRUE
WHERE
    TYPE = 'SLACK_DM'
    AND NOT enable_status_updates;

-- name: StatusMgrCleanupDisabledSubs :exec
DELETE FROM alert_status_subscriptions sub USING user_contact_methods cm
WHERE sub.contact_method_id = cm.id
    AND (cm.disabled
        OR NOT cm.enable_status_updates);

-- name: StatusMgrNextUpdate :one
SELECT
    sub.id,
    channel_id,
    contact_method_id,
    alert_id,
(
        SELECT
            status
        FROM
            alerts a
        WHERE
            a.id = sub.alert_id)
FROM
    alert_status_subscriptions sub
WHERE
    sub.last_alert_status !=(
        SELECT
            status
        FROM
            alerts a
        WHERE
            a.id = sub.alert_id)
LIMIT 1
FOR UPDATE
    SKIP LOCKED;

-- name: StatusMgrLogEntry :one
SELECT
    id,
    sub_user_id AS user_id
FROM
    alert_logs
WHERE
    alert_id = @alert_id::bigint
    AND event = @event_type::enum_alert_log_event
    AND timestamp > now() - '1 hour'::interval
ORDER BY
    id DESC
LIMIT 1;

-- name: StatusMgrCMInfo :one
SELECT
    user_id,
    type
FROM
    user_contact_methods
WHERE
    id = $1
    AND NOT disabled
    AND enable_status_updates;

-- name: StatusMgrDeleteSub :exec
DELETE FROM alert_status_subscriptions
WHERE id = $1;

-- name: StatusMgrSendUserMsg :exec
INSERT INTO outgoing_messages(id, message_type, contact_method_id, user_id, alert_id, alert_log_id)
    VALUES (@id::uuid, 'alert_status_update', @cm_id::uuid, @user_id::uuid, @alert_id::bigint, @log_id);

-- name: StatusMgrSendChannelMsg :exec
INSERT INTO outgoing_messages(id, message_type, channel_id, alert_id, alert_log_id)
    VALUES (@id::uuid, 'alert_status_update', @channel_id::uuid, @alert_id::bigint, @log_id);

-- name: StatusMgrUpdateSub :exec
UPDATE
    alert_status_subscriptions
SET
    last_alert_status = $2
WHERE
    id = $1;

-- name: StatusMgrCleanupStaleSubs :exec
DELETE FROM alert_status_subscriptions sub
WHERE sub.updated_at < now() - '7 days'::interval;

