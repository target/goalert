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

