
-- +migrate Up
ALTER TABLE user_notification_rules RENAME COLUMN delay TO delay_minutes;

-- +migrate Down
ALTER TABLE user_notification_rules RENAME COLUMN delay_minutes TO delay;
