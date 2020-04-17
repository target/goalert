-- +migrate Up notransaction

ALTER TYPE enum_alert_log_event ADD VALUE IF NOT EXISTS 'no_notification_sent';
ALTER TYPE enum_alert_log_subject_type ADD VALUE IF NOT EXISTS 'no_notification';

-- +migrate Down

