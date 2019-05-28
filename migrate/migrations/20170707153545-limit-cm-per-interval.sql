
-- +migrate Up
ALTER TABLE user_notification_rules ADD UNIQUE(contact_method_id, delay_minutes);

-- +migrate Down
