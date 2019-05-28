
-- +migrate Up notransaction

ALTER TYPE enum_alert_log_event ADD VALUE IF NOT EXISTS 'acknowledged';
ALTER TYPE enum_alert_log_event ADD VALUE IF NOT EXISTS 'policy_updated';
ALTER TYPE enum_alert_log_event ADD VALUE IF NOT EXISTS 'duplicate_suppressed';
ALTER TYPE enum_alert_log_event ADD VALUE IF NOT EXISTS 'escalation_request';

-- +migrate Down
