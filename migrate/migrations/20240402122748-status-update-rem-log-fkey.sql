-- +migrate Up
ALTER TABLE outgoing_messages
    DROP CONSTRAINT IF EXISTS outgoing_messages_alert_log_id_fkey;

-- +migrate Down
ALTER TABLE outgoing_messages
    ADD CONSTRAINT outgoing_messages_alert_log_id_fkey FOREIGN KEY (alert_log_id) REFERENCES alert_logs(id) ON DELETE CASCADE;

