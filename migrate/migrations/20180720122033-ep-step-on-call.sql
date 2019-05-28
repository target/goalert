
-- +migrate Up

CREATE TABLE ep_step_on_call_users (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    ep_step_id UUID NOT NULL REFERENCES escalation_policy_steps (id) ON DELETE CASCADE,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    end_time TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX idx_ep_step_on_call
ON ep_step_on_call_users (user_id, ep_step_id)
WHERE end_time IS NULL;

-- +migrate Down

DROP TABLE ep_step_on_call_users;
