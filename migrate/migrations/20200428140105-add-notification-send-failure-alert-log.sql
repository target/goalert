-- +migrate Up notransaction

ALTER TYPE enum_alert_log_event ADD VALUE IF NOT EXISTS 'notification_send_failure';

-- +migrate Down

