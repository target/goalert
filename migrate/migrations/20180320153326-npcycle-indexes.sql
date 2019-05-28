
-- +migrate Up

CREATE INDEX idx_notif_rule_creation_time on user_notification_rules (user_id, created_at);
CREATE INDEX idx_outgoing_messages_notif_cycle on outgoing_messages (cycle_id);
ALTER TABLE notification_policy_cycles SET (fillfactor = 65);
ALTER TABLE outgoing_messages SET (fillfactor = 85);

-- +migrate Down

DROP INDEX idx_notif_rule_creation_time;
DROP INDEX idx_outgoing_messages_notif_cycle;
ALTER TABLE notification_policy_cycles RESET (fillfactor);
ALTER TABLE outgoing_messages RESET (fillfactor);
