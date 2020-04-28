-- +migrate Up notransaction

ALTER TYPE enum_alert_log_event ADD VALUE IF NOT EXISTS 'notification_send_failure';

ALTER TABLE alert_logs
    ADD COLUMN sub_suffix TEXT NOT NULL DEFAULT '';

-- +migrate Down

ALTER TABLE alert_logs
    DROP COLUMN sub_suffix;