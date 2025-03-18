-- name: Alert_LockOneAlertService :one
-- Locks the service associated with the alert.
SELECT
    maintenance_expires_at NOTNULL::bool AS is_maint_mode,
    alerts.status
FROM
    services svc
    JOIN alerts ON alerts.service_id = svc.id
WHERE
    alerts.id = $1
FOR UPDATE;

-- name: Alert_RequestAlertEscalationByTime :one
-- Returns the alert ID and the escalation policy ID for the alert that should be escalated based on the provided time.
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

-- name: Alert_AlertHasEPState :one
-- Returns true if the alert has an escalation policy state.
SELECT
    EXISTS (
        SELECT
            1
        FROM
            escalation_policy_state
        WHERE
            alert_id = $1) AS has_ep_state;

-- name: Alert_GetAlertFeedback :many
-- Returns the noise reason for the alert.
SELECT
    alert_id,
    noise_reason
FROM
    alert_feedback
WHERE
    alert_id = ANY ($1::int[]);

-- name: Alert_SetAlertFeedback :exec
-- Sets the noise reason for the alert.
INSERT INTO alert_feedback(alert_id, noise_reason)
    VALUES ($1, $2)
ON CONFLICT (alert_id)
    DO UPDATE SET
        noise_reason = $2
    WHERE
        alert_feedback.alert_id = $1;

-- name: Alert_SetManyAlertFeedback :many
-- Sets the noise reason for many alerts.
INSERT INTO alert_feedback(alert_id, noise_reason)
    VALUES (unnest(@alert_ids::bigint[]), @noise_reason)
ON CONFLICT (alert_id)
    DO UPDATE SET
        noise_reason = excluded.noise_reason
    WHERE
        alert_feedback.alert_id = excluded.alert_id
    RETURNING
        alert_id;

-- name: Alert_GetAlertMetadata :one
-- Returns the metadata for the alert.
SELECT
    metadata
FROM
    alert_data
WHERE
    alert_id = $1;

-- name: Alert_GetAlertManyMetadata :many
-- Returns the metadata for many alerts.
SELECT
    alert_id,
    metadata
FROM
    alert_data
WHERE
    alert_id = ANY (@alert_ids::bigint[]);

-- name: Alert_SetAlertMetadata :execrows
-- Sets the metadata for the alert.
INSERT INTO alert_data(alert_id, metadata)
SELECT
    a.id,
    $2
FROM
    alerts a
WHERE
    a.id = $1
    AND a.status != 'closed'
    AND (a.service_id = $3
        OR $3 IS NULL) -- ensure the alert is associated with the service, if coming from an integration
ON CONFLICT (alert_id)
    DO UPDATE SET
        metadata = $2
    WHERE
        alert_data.alert_id = $1;

-- name: Alert_ServiceEPHasSteps :one
-- Returns true if the Escalation Policy for the provided service has at least one step.
SELECT coalesce(
    (SELECT true
    FROM escalation_policies pol
    JOIN services svc ON svc.id = @id::text
    WHERE
        pol.id = svc.escalation_policy_id
        AND pol.step_count = 0)
, false);

-- name: Alert_LockService :exec
-- Locks the service associated with the alert.
SELECT 1
FROM
    services
WHERE
    id = @id::text
FOR
UPDATE;

-- name: Alert_LockAlertService :exec
-- Locks the alert service associated with the alert.
SELECT 1
FROM services s
JOIN alerts a ON a.id = ANY (@alert_ids::bigint[])
AND s.id = a.service_id
FOR
UPDATE;

-- name: Alert_GetStatusAndLockService :one
-- Returns the status of the alert and locks the service associated with the alert.
SELECT a.status
FROM services s
JOIN alerts a ON a.id = @id::bigint
AND a.service_id = s.id
FOR
UPDATE;

-- name: Alert_GetEscalationPolicyID :one
-- Returns the escalation policy ID associated with the alert.
SELECT escalation_policy_id
FROM
    services svc,
    alerts a
WHERE
    svc.id = @id::bigint;
