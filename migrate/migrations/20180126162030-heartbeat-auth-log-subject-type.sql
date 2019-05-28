
-- +migrate Up notransaction

ALTER TYPE enum_alert_log_subject_type ADD VALUE IF NOT EXISTS 'heartbeat_monitor';

-- +migrate Down
