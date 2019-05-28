
-- +migrate Up notransaction

ALTER TYPE enum_alert_source ADD VALUE IF NOT EXISTS 'email';

-- +migrate Down
