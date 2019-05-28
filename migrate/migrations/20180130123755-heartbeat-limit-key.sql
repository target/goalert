
-- +migrate Up notransaction
ALTER TYPE enum_limit_type ADD VALUE IF NOT EXISTS 'heartbeat_monitors_per_service';

-- +migrate Down
