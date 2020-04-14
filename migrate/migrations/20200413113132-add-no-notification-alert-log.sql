-- +migrate Up notransaction

ALTER TYPE enum_alert_log_event ADD VALUE IF NOT EXISTS 'no_notification_sent';

-- +migrate Down

