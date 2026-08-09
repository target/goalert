-- +migrate Up notransaction
-- Add new integration key type 'azureMonitor'

ALTER TYPE enum_integration_keys_type ADD VALUE IF NOT EXISTS 'azureMonitor';
ALTER TYPE enum_alert_source ADD VALUE IF NOT EXISTS 'azureMonitor';

-- +migrate Down
