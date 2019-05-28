
-- +migrate Up

CREATE TABLE escalation_policy_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    escalation_policy_step_id TEXT NOT NULL REFERENCES escalation_policy_step (id) ON DELETE CASCADE,
    schedule_id TEXT REFERENCES schedule (id) ON DELETE CASCADE,
    user_id UUID REFERENCES users (id) ON DELETE CASCADE,

    UNIQUE(escalation_policy_step_id, schedule_id, user_id),
    CHECK((schedule_id IS NOT NULL AND user_id IS NULL) OR (user_id IS NOT NULL AND schedule_id IS NULL))
);

INSERT INTO escalation_policy_actions (id, escalation_policy_step_id, schedule_id, user_id)
    SELECT id::UUID, escalation_policy_step_id, 
        CASE WHEN type_text = 'schedule_reference' THEN type_id ELSE NULL END,
        CASE WHEN type_text = 'user_reference' THEN type_id::UUID ELSE NULL END
    FROM escalation_policy_action;

DROP TABLE escalation_policy_action;

-- +migrate Down

CREATE TABLE escalation_policy_action (
	id TEXT PRIMARY KEY,
	escalation_policy_step_id TEXT REFERENCES escalation_policy_step(id),
  type_id TEXT, --user or schedule id
  type_text TEXT --can be user_reference or schedule_reference
);

INSERT INTO escalation_policy_action (id, escalation_policy_step_id, type_id, type_text)
    SELECT id::TEXT, escalation_policy_step_id,
        CASE WHEN schedule_id IS NULL THEN user_id::TEXT ELSE schedule_id END,
        CASE WHEN schedule_id IS NULL THEN 'user_reference' ELSE 'schedule_reference' END
    FROM escalation_policy_actions;

DROP TABLE escalation_policy_actions;
