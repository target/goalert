
-- +migrate Up
ALTER TABLE escalation_policy_state
    ALTER CONSTRAINT svc_ep_fkey DEFERRABLE;

-- +migrate Down
ALTER TABLE escalation_policy_state
    ALTER CONSTRAINT svc_ep_fkey NOT DEFERRABLE;
