-- name: LockOneAlertService :one
SELECT
    maintenance_expires_at NOTNULL::bool AS is_maint_mode,
    alerts.status
FROM
    services svc
    JOIN alerts ON alerts.service_id = svc.id
WHERE
    alerts.id = $1
FOR UPDATE;

-- name: RequestAlertEscalationByTime :one
UPDATE
    escalation_policy_state
SET
    force_escalation = TRUE
WHERE
    alert_id = $1
    AND (last_escalation <= $2::timestamptz
        OR last_escalation IS NULL)
RETURNING
    TRUE;

-- name: AlertHasEPState :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            escalation_policy_state
        WHERE
            alert_id = $1) AS has_ep_state;

-- name: AlertFeedback :many
SELECT
    alert_id,
    noise_reason
FROM
    alert_feedback
WHERE
    alert_id = ANY($1::int[]);

-- name: SetAlertFeedback :exec
INSERT INTO alert_feedback(alert_id, noise_reason)
    VALUES ($1, $2)
ON CONFLICT (alert_id)
    DO UPDATE SET
        noise_reason = $2
    WHERE
        alert_feedback.alert_id = $1;

-- name: SetManyAlertFeedback :many
INSERT INTO alert_feedback(alert_id, noise_reason)
VALUES (unnest($1::bigint[]), $2)
ON CONFLICT (alert_id)
    DO UPDATE SET
        noise_reason = excluded.noise_reason
    WHERE
        alert_feedback.alert_id = excluded.alert_id
RETURNING alert_id;
