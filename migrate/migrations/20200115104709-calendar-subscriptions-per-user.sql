
-- +migrate Up
ALTER TYPE enum_limit_type ADD VALUE IF NOT EXISTS 'calendar_subscriptions_per_user';

INSERT INTO config_limits (id, max)
VALUES
	('calendar_subscriptions_per_user', 15);

-- +migrate Down
