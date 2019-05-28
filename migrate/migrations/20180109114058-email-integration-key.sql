
-- +migrate Up notransaction

ALTER TYPE enum_integration_keys_type ADD VALUE IF NOT EXISTS 'email';

-- +migrate Down
