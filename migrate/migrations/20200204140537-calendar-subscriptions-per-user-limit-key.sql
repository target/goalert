
-- +migrate Up notransaction
ALTER TYPE enum_limit_type ADD VALUE IF NOT EXISTS 'calendar_subscriptions_per_user';

-- +migrate Down
