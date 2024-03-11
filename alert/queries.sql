-- name: Insert :one
INSERT INTO alerts(summary, details, service_id, source, status, dedup_key)
    VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    id, created_at
    -- name: NoStepsByService :one
    SELECT
        coalesce((
            SELECT
                TRUE
            FROM escalation_policies pol
            JOIN services svc ON svc.id = $1
            WHERE
                pol.id = svc.escalation_policy_id
                AND pol.step_count = 0), FALSE)

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
    alert_id = ANY ($1::int[]);

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
    VALUES (unnest(@alert_ids::bigint[]), @noise_reason)
ON CONFLICT (alert_id)
    DO UPDATE SET
        noise_reason = excluded.noise_reason
    WHERE
        alert_feedback.alert_id = excluded.alert_id
    RETURNING
        alert_id;

-- name: AlertMetadata :one
SELECT
    alert_id,
    metadata
FROM
    alerts_metadata
WHERE
    alert_id = ANY ($1::number);

-- name: SetAlertMetadata :exec
INSERT INTO alerts_metadata(alert_id, metadata)
    VALUES ($1, $2)
ON CONFLICT (alert_id)
    DO UPDATE SET
        metadata = $2
    WHERE
        alerts_metadata.alert_id = $1;

