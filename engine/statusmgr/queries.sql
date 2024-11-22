-- name: StatusMgrOutdated :many
SELECT
    sub.id
FROM
    alert_status_subscriptions sub
    JOIN alerts a ON a.id = sub.alert_id
WHERE
    sub.last_alert_status != a.status;

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
    last_alert_status = $2,
    updated_at = now()
WHERE
    id = $1;

-- name: StatusMgrCleanupStaleSubs :exec
DELETE FROM alert_status_subscriptions sub
WHERE sub.updated_at < now() - '7 days'::interval;

-- name: StatusMgrFindOne :one
SELECT
    sub.*,
    a.status
FROM
    alert_status_subscriptions sub
    JOIN alerts a ON a.id = sub.alert_id
WHERE
    sub.id = $1
FOR UPDATE
    SKIP LOCKED;

