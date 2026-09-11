-- +migrate Up

-- Multi-ack: when enabled on a step, notification cycles started by that step
-- continue after the alert is acknowledged, so that everyone on the step is
-- notified and can acknowledge.
ALTER TABLE escalation_policy_steps
    ADD COLUMN multi_ack boolean NOT NULL DEFAULT false;

-- Stamped onto each cycle when it is created, so a cycle's behavior is fixed at
-- creation time and doesn't change if the escalation policy is edited later.
ALTER TABLE notification_policy_cycles
    ADD COLUMN multi_ack boolean NOT NULL DEFAULT false;

-- +migrate Down

ALTER TABLE escalation_policy_steps
    DROP COLUMN multi_ack;

ALTER TABLE notification_policy_cycles
    DROP COLUMN multi_ack;
