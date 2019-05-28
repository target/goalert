
-- +migrate Up

ALTER TABLE escalation_policy_actions
    ADD COLUMN rotation_id UUID REFERENCES rotations (id) ON DELETE CASCADE,

    DROP CONSTRAINT escalation_policy_actions_escalation_policy_step_id_schedul_key,
    DROP CONSTRAINT escalation_policy_actions_check,

    ADD CONSTRAINT epa_no_duplicate_users UNIQUE(escalation_policy_step_id, user_id),
    ADD CONSTRAINT epa_no_duplicate_schedules UNIQUE(escalation_policy_step_id, schedule_id),
    ADD CONSTRAINT epa_no_duplicate_rotations UNIQUE(escalation_policy_step_id, rotation_id),
    ADD CONSTRAINT epa_there_can_only_be_one CHECK (
        (user_id IS NULL AND schedule_id IS NULL AND rotation_id IS NOT NULL)
        OR
        (user_id IS NULL AND schedule_id IS NOT NULL AND rotation_id IS NULL)
        OR
        (user_id IS NOT NULL AND schedule_id IS NULL AND rotation_id IS NULL)
    );


-- +migrate Down

DELETE FROM escalation_policy_actions WHERE rotation_id IS NOT NULL;

ALTER TABLE escalation_policy_actions
    DROP COLUMN rotation_id,
    DROP CONSTRAINT epa_no_duplicate_schedules,
    DROP CONSTRAINT epa_no_duplicate_users,
    ADD CONSTRAINT escalation_policy_actions_escalation_policy_step_id_schedul_key UNIQUE(escalation_policy_step_id, schedule_id, user_id),
    ADD CONSTRAINT escalation_policy_actions_check CHECK (
        (schedule_id IS NOT NULL AND user_id IS NULL)
        OR
        (user_id IS NOT NULL AND schedule_id IS NULL)
    );
