
-- +migrate Up
-- Add new alert_log_event type
ALTER TYPE enum_alert_log_event RENAME TO enum_alert_log_event_old;
CREATE TYPE enum_alert_log_event AS ENUM (
	'created',
	'reopened',
    'status_changed',
    'assignment_changed',
    'escalated',
    'closed',
    'notification_sent',
    'response_received'
);
ALTER TABLE alert_logs ALTER COLUMN event TYPE enum_alert_log_event USING event::TEXT::enum_alert_log_event;
DROP TYPE enum_alert_log_event_old;

-- +migrate Down

ALTER TYPE enum_alert_log_event RENAME TO enum_alert_log_event_old;
CREATE TYPE enum_alert_log_event AS ENUM (
	'created',
	'reopened',
    'status_changed',
    'assignment_changed',
    'escalated',
    'closed'
);

ALTER TABLE alert_logs ALTER COLUMN event TYPE enum_alert_log_event USING event::TEXT::enum_alert_log_event;
DROP TYPE enum_alert_log_event_old;
