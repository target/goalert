-- +migrate Up notransaction
-- Add new integration key type 'cloudwatch'

ALTER TYPE enum_integration_keys_type ADD VALUE IF NOT EXISTS 'cloudwatch';
ALTER TYPE enum_alert_source ADD VALUE IF NOT EXISTS 'cloudwatch';

-- +migrate Down
