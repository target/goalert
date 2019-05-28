
-- +migrate Up
ALTER TABLE escalation_policy_state
    ADD COLUMN next_escalation TIMESTAMP WITH TIME ZONE;

CREATE INDEX ON escalation_policy_state (next_escalation, force_escalation);

-- +migrate Down

ALTER TABLE escalation_policy_state
    DROP COLUMN next_escalation;

