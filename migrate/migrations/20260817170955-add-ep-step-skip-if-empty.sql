-- +migrate Up

-- When set, an escalation step that resolves to no one -- no on-call users and
-- no notification channels -- escalates immediately instead of consuming its
-- delay. Waiting on such a step gains nothing: on-call is resolved once, at
-- escalation time, so a user whose shift begins partway through the delay is
-- not notified until the following escalation anyway.
ALTER TABLE escalation_policy_steps
    ADD COLUMN skip_if_empty BOOLEAN NOT NULL DEFAULT FALSE;

-- +migrate Down

ALTER TABLE escalation_policy_steps
    DROP COLUMN skip_if_empty;
