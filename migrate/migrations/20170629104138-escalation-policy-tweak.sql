
-- +migrate Up
ALTER TABLE escalation_policy
    ALTER COLUMN id SET DEFAULT gen_random_uuid()::TEXT;

-- +migrate Down
