
-- +migrate Up notransaction
-- Add new integration key type 'prometheusAlertmanager'

ALTER TYPE enum_integration_keys_type ADD VALUE IF NOT EXISTS 'prometheusAlertmanager';
ALTER TYPE enum_alert_source ADD VALUE IF NOT EXISTS 'prometheusAlertmanager';

-- +migrate Down
