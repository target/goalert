
-- +migrate Up notransaction
ALTER TYPE enum_limit_type ADD VALUE IF NOT EXISTS 'user_overrides_per_schedule';

-- +migrate Down
