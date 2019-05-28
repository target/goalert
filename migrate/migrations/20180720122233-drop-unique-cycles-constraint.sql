
-- +migrate Up
ALTER TABLE notification_policy_cycles
    DROP CONSTRAINT notification_policy_cycles_user_id_alert_id_key;

-- +migrate Down

ALTER TABLE notification_policy_cycles
    ADD CONSTRAINT notification_policy_cycles_user_id_alert_id_key UNIQUE (user_id, alert_id);
