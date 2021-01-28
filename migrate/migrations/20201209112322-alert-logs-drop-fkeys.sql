-- +migrate Up

ALTER TABLE alert_logs
    DROP CONSTRAINT alert_logs_alert_id_fkey,
    DROP CONSTRAINT alert_logs_sub_user_id_fkey,
    DROP CONSTRAINT alert_logs_sub_integration_key_id_fkey,
    DROP CONSTRAINT alert_logs_sub_hb_monitor_id_fkey,
    DROP CONSTRAINT alert_logs_sub_channel_id_fkey;

-- +migrate Down

ALTER TABLE alert_logs
    ADD CONSTRAINT alert_logs_alert_id_fkey FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE,
    ADD CONSTRAINT alert_logs_sub_user_id_fkey FOREIGN KEY (sub_user_id) REFERENCES users(id) ON DELETE SET NULL,
    ADD CONSTRAINT alert_logs_sub_integration_key_id_fkey FOREIGN KEY (sub_integration_key_id) REFERENCES integration_keys(id) ON DELETE SET NULL,
    ADD CONSTRAINT alert_logs_sub_hb_monitor_id_fkey FOREIGN KEY (sub_hb_monitor_id) REFERENCES heartbeat_monitors(id) ON DELETE SET NULL,
    ADD CONSTRAINT alert_logs_sub_channel_id_fkey FOREIGN KEY (sub_channel_id) REFERENCES notification_channels(id) ON DELETE SET NULL;
