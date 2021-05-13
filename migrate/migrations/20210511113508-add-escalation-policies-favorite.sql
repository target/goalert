-- +migrate Up
ALTER TABLE user_favorites
    ADD COLUMN tgt_escalation_policy_id UUID REFERENCES escalation_policies(id) ON DELETE CASCADE,
    ADD CONSTRAINT user_favorites_user_id_tgt_escalation_policies_key UNIQUE(user_id, tgt_escalation_policy_id);
-- +migrate Down
ALTER TABLE user_favorites
    DROP CONSTRAINT user_favorites_user_id_tgt_escalation_policies_id_key,
    DROP COLUMN tgt_escalation_policy_id;