
-- +migrate Up


ALTER TABLE escalation_policy_step
DROP CONSTRAINT escalation_policy_step_escalation_policy_id_fkey,
ADD CONSTRAINT escalation_policy_step_escalation_policy_id_fkey
	FOREIGN KEY (escalation_policy_id)
	REFERENCES escalation_policy(id)
	ON DELETE CASCADE;

-- +migrate Down
