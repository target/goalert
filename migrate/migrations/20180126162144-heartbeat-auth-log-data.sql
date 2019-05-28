
-- +migrate Up

ALTER TABLE alert_logs
    ADD COLUMN sub_hb_monitor_id UUID REFERENCES heartbeat_monitors (id) ON DELETE SET NULL,
    DROP CONSTRAINT alert_logs_one_subject,
    ADD CONSTRAINT alert_logs_one_subject CHECK(
        NOT (sub_user_id NOTNULL AND sub_integration_key_id NOTNULL AND sub_hb_monitor_id NOTNULL)
    )
;

-- +migrate Down

ALTER TABLE alert_logs
    DROP COLUMN sub_hb_monitor_id,
    ADD CONSTRAINT alert_logs_one_subject CHECK(
        NOT (sub_user_id NOTNULL AND sub_integration_key_id NOTNULL)
    )
;
