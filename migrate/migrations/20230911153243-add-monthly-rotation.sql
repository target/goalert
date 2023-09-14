-- +migrate Up notransaction
ALTER TYPE enum_rotation_type ADD VALUE IF NOT EXISTS 'monthly';

-- +migrate Down

