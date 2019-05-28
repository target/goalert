
-- +migrate Up

ALTER TABLE escalation_policy_actions
    ADD COLUMN channel_id UUID REFERENCES notification_channels (id) ON DELETE CASCADE,
    DROP CONSTRAINT epa_there_can_only_be_one,
    ADD CONSTRAINT epa_there_can_only_be_one CHECK (
        (case when user_id notnull then 1 else 0 end +
        case when schedule_id notnull then 1 else 0 end +
        case when rotation_id notnull then 1 else 0 end +
        case when channel_id notnull then 1 else 0 end) = 1
    ),
    ADD CONSTRAINT epa_no_duplicate_channels UNIQUE (escalation_policy_step_id, channel_id);

-- +migrate Down

ALTER TABLE escalation_policy_actions
    DROP COLUMN channel_id,
    ADD CONSTRAINT epa_there_can_only_be_one CHECK (user_id IS NULL AND schedule_id IS NULL AND rotation_id IS NOT NULL OR user_id IS NULL AND schedule_id IS NOT NULL AND rotation_id IS NULL OR user_id IS NOT NULL AND schedule_id IS NULL AND rotation_id IS NULL);
