-- +migrate Up

ALTER TYPE enum_integration_keys_type ADD VALUE IF NOT EXISTS 'universal';
ALTER TYPE enum_alert_source ADD VALUE IF NOT EXISTS 'universal';

-- +migrate Down


