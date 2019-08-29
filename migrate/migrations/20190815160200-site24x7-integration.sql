
-- +migrate Up notransaction
-- Add new integration key type 'site24x7'

ALTER TYPE enum_integration_keys_type ADD VALUE IF NOT EXISTS 'site24x7';
ALTER TYPE enum_alert_source ADD VALUE IF NOT EXISTS 'site24x7';

-- +migrate Down
