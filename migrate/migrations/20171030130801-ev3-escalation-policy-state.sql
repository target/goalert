
-- +migrate Up
CREATE TABLE escalation_policy_state (
    escalation_policy_id UUID NOT NULL REFERENCES escalation_policies (id) ON DELETE CASCADE,
    escalation_policy_step_id UUID REFERENCES escalation_policy_steps (id) ON DELETE SET NULL,
    escalation_policy_step_number INT NOT NULL DEFAULT 0,
    alert_id BIGINT NOT NULL REFERENCES alerts (id) ON DELETE CASCADE,
    last_escalation TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    loop_count INT NOT NULL DEFAULT 0,
    force_escalation BOOLEAN NOT NULL DEFAULT false,

    UNIQUE(alert_id, escalation_policy_id)
);

WITH ep_step_count AS (
    SELECT count(id) as max, escalation_policy_id
    FROM escalation_policy_steps
    GROUP BY escalation_policy_id
)
INSERT INTO escalation_policy_state (
    escalation_policy_id,
    escalation_policy_step_id,
    alert_id,
    last_escalation,
    loop_count
)
SELECT
    svc.escalation_policy_id,
    step.id,
    alert.id,
    alert.last_escalation,
    alert.escalation_level / cnt.max
FROM
    alerts alert,
    services svc,
    ep_step_count cnt,
    escalation_policy_steps step
WHERE svc.id = alert.service_id
    AND step.escalation_policy_id = svc.escalation_policy_id
    AND cnt.escalation_policy_id = svc.escalation_policy_id
    AND step.step_number = alert.escalation_level % cnt.max;

-- +migrate Down

WITH ep_step_count AS (
    SELECT count(id) as max, escalation_policy_id
    FROM escalation_policy_steps
    GROUP BY escalation_policy_id
)
UPDATE alerts
SET
    escalation_level = cnt.max * state.loop_count + step.step_number,
    last_escalation = state.last_escalation
FROM
    escalation_policy_state state,
    escalation_policy_steps step,
    ep_step_count cnt
WHERE step.id = state.escalation_policy_step_id
    AND alerts.id = state.alert_id
    AND cnt.escalation_policy_id = state.escalation_policy_id;    

DROP TABLE escalation_policy_state;
